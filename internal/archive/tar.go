package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	"github.com/internet-equity/traceneck/internal/network"
)

func Write() {
	var outFile *os.File

	if config.OutPath == "-" {
		outFile = os.Stdout
	} else {
		var err error
		outFile, err = os.Create(config.OutPath)
		if err != nil {
			log.Println("[archive] error opening archive file:", err)
			return
		}
		defer outFile.Close()
	}

	var archive *tar.Writer

	if outExt := filepath.Ext(config.OutPath); outExt == ".gz" || outExt == ".tgz" {
		writer, _ := gzip.NewWriterLevel(outFile, gzip.BestCompression)
		defer writer.Close()

		archive = tar.NewWriter(writer)
	} else {
		archive = tar.NewWriter(outFile)
	}
	defer archive.Close()

	if err := addFileToTar(archive, meta.MetaFile); err != nil {
		log.Fatalln("[archive] error writing", meta.MetaFile, "to tar:", err.Error())
	}

	if err := addFileToTar(archive, network.CapFile); err != nil {
		log.Fatalln("[archive] error writing", network.CapFile, "to tar:", err.Error())
	}

	log.Println("[archive] data archived to:", config.OutPath)
}

func addFileToTar(archive *tar.Writer, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if err := archive.WriteHeader(&tar.Header{
		Name:    filepath.Base(fileName),
		Size:    fileInfo.Size(),
		Mode:    int64(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
	}); err != nil {
		return err
	}

	_, err = io.Copy(archive, file)
	return err
}
