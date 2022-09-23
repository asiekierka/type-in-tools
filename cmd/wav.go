/*
Copyright © 2022 Adrian 'asie' Siekierka
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/asiekierka/fbastool/internal"
	"github.com/spf13/cobra"
)

// wavCmd represents the wav command
var wavCmd = &cobra.Command{
	Use:   "wav",
	Short: "Convert WAV files to/from binary data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runWav(args[0])
	},
}

func runWav(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	tapeEncInfo := internal.NewTapeEncodingInfo()
	tapeReader, err := internal.NewTapeReader(fp, tapeEncInfo)
	if err != nil {
		panic(err)
	}

	for {
		blockType, err := tapeReader.SyncToBlock()
		if err != nil {
			fmt.Println("block sync error:", err)
			break
		}
		fmt.Println(blockType)

		err = tapeReader.VerifyBit(1)
		if err != nil {
			fmt.Println("block prelude error:", err)
			break
		}

		blockData, err := tapeReader.NextBytes(128)
		if err != nil {
			fmt.Println("block read error:", err)
			break
		}
		fmt.Println(blockData)
	}

	/* pulseLast := uint8(0)
	pulseCount := 0

	for {
		pulseBit, err := tapeReader.NextBit()
		if err != nil {
			break
		}
		if pulseBit == pulseLast {
			pulseCount++
		} else {
			fmt.Printf("read %d x %d\n", pulseBit, pulseCount+1)
			pulseLast = pulseBit
			pulseCount = 0
		}
	} */
}

func init() {
	rootCmd.AddCommand(wavCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wavCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wavCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
