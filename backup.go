package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BackupOptions struct {
	RootDir   string
	WorldDir  string
	BackupDir string
	Interval  time.Duration
	PruneTime time.Duration
}

func backup(server *Server, opts BackupOptions) {
	t := time.NewTicker(opts.Interval)

	for {
		select {
		case <-t.C:
			{
				server.Write(`tellraw @p ["§f[§6CakeMC§f] ",{"text":"Starting backup of server...","color":"purple"}]`)
				err := prune(opts)
				if err != nil {
					log.Println("Prune failed: ", err)
					server.Write(`tellraw @p ["§f[§6CakeMC§f] ",{"text":"Error saving! Check logs for details.","color":"red", "bold":true}]`)
					return
				}
				server.Write("save-off")
				name, err := archive(opts)
				if err != nil {
					log.Println("Archive failed: ", err)
					server.Write(`tellraw @p ["§f[§6CakeMC§f] ""§f[§6CakeMC§f] ",{"text":"Error saving! Check logs for details.","color":"red", "bold":true}]`)
					return
				}
				server.Write("save-on")
				server.Write(`tellraw @p ["§f[§6CakeMC§f] ",{"text":"Backup completed!","color":"purple","hoverEvent":{"action":"show_text", "value":["` + name + `"]}}]`)
				log.Println("Backed up", opts.WorldDir, ":", name)
			}

		}
	}
}

func prune(opts BackupOptions) error {
	return filepath.Walk(opts.BackupDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".tar.gz") && info.ModTime().Sub(time.Now()) < -opts.PruneTime {
			if err := os.Remove(opts.BackupDir + "/" + info.Name()); err != nil {
				return err
			}
		}

		return nil
	})
}

func archive(opts BackupOptions) (string, error) {
	var buff bytes.Buffer

	zr := gzip.NewWriter(&buff)
	tw := tar.NewWriter(zr)

	filepath.Walk(opts.RootDir+"/"+opts.WorldDir, func(path string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(path)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}

			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}

		return nil
	})

	if err := tw.Close(); err != nil {
		return "", err
	}

	if err := zr.Close(); err != nil {
		return "", err
	}

	name := opts.WorldDir + "_" + time.Now().Format(time.RFC3339) + ".tar.gz"

	file, err := os.OpenFile(opts.BackupDir+"/"+name, os.O_CREATE|os.O_RDWR, os.FileMode(600))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(file, &buff); err != nil {
		return "", err
	}

	return name, nil
}
