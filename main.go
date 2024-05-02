package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

const SPEECH_LANGUAGE_JP = "ja_JP"
const SPEECH_LANGUAGE_EN = "en_US"

func main() {
	bytes, err := os.ReadFile("text.csv")
	if err != nil {
		log.Fatal(err.Error())
	}

	_, rows, err := ReadFromBytes(bytes)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var wg sync.WaitGroup
	wg.Add(len(rows) * 2)

	for i, row := range rows {
		commonFileName := fmt.Sprintf("outputs/%d_%s", i, row["english"])
		go func() {
			filename := commonFileName + "_" + SPEECH_LANGUAGE_JP + ".mp3"
			fmt.Println(filename)
			// if err := GenerateAiSpeech(ctx, client, row["japanese"], SPEECH_LANGUAGE_JP, filename); err != nil {
			// 	log.Fatal(err)
			// }
			wg.Done()
		}()

		go func() {
			filename := commonFileName + "_" + SPEECH_LANGUAGE_EN + ".mp3"
			fmt.Println(filename)
			// if err := GenerateAiSpeech(ctx, client, row["english"], SPEECH_LANGUAGE_EN, filename); err != nil {
			// 	log.Fatal(err)
			// }
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("出力完了")
}

func ReadFromBytes(data []byte) (headers []string, rows []map[string]string, err error) {
	// BOMがファイルの先頭にあれば削除する
	urf8BOM := []byte{239, 187, 191}
	bomTrimmedData := bytes.TrimPrefix(data, urf8BOM)

	rd := csv.NewReader(bytes.NewBuffer(bomTrimmedData))
	rows = []map[string]string{}
	for {
		record, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return headers, rows, err
		}
		if headers == nil {
			headers = record
		} else {
			dict := map[string]string{}
			for i := range headers {
				dict[headers[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return headers, rows, nil
}

func GenerateAiSpeech(ctx context.Context, client *texttospeech.Client, text string, language string, filename string) error {
	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: language,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	// The resp's AudioContent is binary.
	err = os.WriteFile(filename, resp.AudioContent, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Audio content written to file: %v\n", filename)

	return nil
}
