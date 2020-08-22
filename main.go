package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/AhmadWaleed/dropbox-uploader/cmd"
)

type chunk struct {
	BufferSize int
	Offset     int64
}

func main() {
	cmd.Execute()
	return

	// test code
	const BufferSize = 5
	file, err := os.Open("test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileinfo, _ := file.Stat()

	filesize := int(fileinfo.Size())

	concurrency := filesize / BufferSize

	chunksizes := make([]chunk, concurrency)

	for i := 0; i < concurrency; i++ {
		chunksizes[i].BufferSize = BufferSize
		chunksizes[i].Offset = int64(BufferSize * i)
	}

	if remainder := filesize % BufferSize; remainder != 0 {
		c := chunk{BufferSize: remainder, Offset: int64(concurrency * BufferSize)}
		concurrency++
		chunksizes = append(chunksizes, c)
	}

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(chunksizes []chunk, i int) {
			defer wg.Done()

			chunk := chunksizes[i]
			buffer := make([]byte, chunk.BufferSize)
			bytesread, _ := file.ReadAt(buffer, chunk.Offset)

			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("bytes read, string(bytestream): ", bytesread)
			fmt.Println("bytestream to string: ", string(buffer))
		}(chunksizes, i)
	}

	wg.Wait()

}
