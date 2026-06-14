package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ArditZubaku/kvx/config"
)

const FilePermUserReadWrite os.FileMode = 0o644

func dumpAllAOF() {
	// Open for writing, create if missing, append data, with standard 0644 permissions
	file, err := os.OpenFile(config.AofFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, FilePermUserReadWrite)
	if err != nil {
		log.Printf("error: %+v", err)

		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing AOF file: %+v", err)
		}
	}()

	log.Println("rewriting AOF file at:", config.AofFile)

	for key, obj := range store {
		dumpKey(file, key, obj)
	}

	log.Println("AOF file rewrite complete")
}

// TODO: support non-kv data structures
// TODO: support sync write.
func dumpKey(file *os.File, key string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")

	_, err := file.Write(Encode(tokens))
	if err != nil {
		log.Printf("error writing to file: %s\n err: %+v", file.Name(), err)

		return
	}
}
