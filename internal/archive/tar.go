package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	"github.com/internet-equity/traceneck/internal/network"
	"github.com/internet-equity/traceneck/internal/util"
)

func CreateArchive() {
	archiveFile := strings.ReplaceAll(
		util.GetFilePath(config.OutDir, "measurement.tar.gz", config.Timestamp),
		":", "-",
	)
	archiveWriter, err := os.Create(archiveFile)
	if err != nil {
		log.Println("[archive] error opening archive file:", err)
		return
	}
	defer archiveWriter.Close()

	gw, _ := gzip.NewWriterLevel(archiveWriter, gzip.BestCompression)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addFileToTar(tw, meta.MetaFile); err != nil {
		log.Fatalln("[archive] error writing", meta.MetaFile, "to tar:", err.Error())
	}

	if err := addFileToTar(tw, network.CapFile); err != nil {
		log.Fatalln("[archive] error writing", meta.MetaFile, "to tar:", err.Error())
	}

	os.Remove(meta.MetaFile)
	os.Remove(network.CapFile)

	log.Println("[archive] data archived to:", archiveFile)
}

func addFileToTar(tw *tar.Writer, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if err := tw.WriteHeader(&tar.Header{
		Name:    filepath.Base(fileName),
		Size:    fileInfo.Size(),
		Mode:    int64(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
	}); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}
