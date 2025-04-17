package yt

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// YtCmd represents the youtube command
var YtCmd = &cobra.Command{
	Use:   "yt",
	Short: "Transcribes video and generates YouTube title/description using Ollama/OpenAI.",
	Long: `Uses ffmpeg to extract audio from a video file, 
sends it to OpenAI Whisper for transcription, 
and then sends the transcription to an Ollama model (like Llama3) 
to generate a compelling YouTube title and description.`,
	Run: runYtCommand,
}

func init() {
	// Flags specific to the 'yt' command
	YtCmd.Flags().StringP("video", "v", "", "Path to the video file (required)")
	YtCmd.Flags().StringP("ollama", "o", "http://10.30.0.133:11434", "Ollama API endpoint URL")
	YtCmd.Flags().StringP("generate-model", "g", "llama3", "Ollama model for generation")

	// Bind flags to Viper (for reading from config/env)
	viper.BindPFlag("yt.video", YtCmd.Flags().Lookup("video"))
	viper.BindPFlag("yt.ollama", YtCmd.Flags().Lookup("ollama"))
	viper.BindPFlag("yt.generate-model", YtCmd.Flags().Lookup("generate-model"))

	// Mark required flags (only video is left as a flag)
	YtCmd.MarkFlagRequired("video")
}

// runYtCommand contains the core application logic for the 'yt' command.
func runYtCommand(cmd *cobra.Command, args []string) {
	// Get flag values directly or via Viper (if bound)
	videoPath := viper.GetString("yt.video") // Get potentially overridden value

	// Resolve to absolute path to handle relative paths correctly
	absoluteVideoPath, err := filepath.Abs(videoPath)
	if err != nil {
		log.Fatalf("Error resolving video path %q: %v", videoPath, err)
	}

	// OpenAI Key (Required from config/env)
	openaiKey := viper.GetString("openai_key") // Get directly from config/env
	if openaiKey == "" {
		log.Fatalf("Error: OpenAI API key is required. Set 'openai_key' in your config file (~/.bmoc.yml) or set the BMOC_OPENAI_KEY environment variable.")
	}

	// Ollama settings (get potentially overridden values)
	ollamaEndpoint := viper.GetString("yt.ollama")
	generateModel := viper.GetString("yt.generate-model")

	fmt.Printf("Processing video: %s\n", absoluteVideoPath)
	fmt.Printf("Using OpenAI for Transcription\n")
	fmt.Printf("Using Ollama endpoint for Generation: %s\n", ollamaEndpoint)
	fmt.Printf("Generation Model: %s\n", generateModel)

	// 1. Extract audio from video
	audioPath, err := extractAudio(absoluteVideoPath)
	if err != nil {
		log.Fatalf("Error extracting audio: %v", err)
	}
	defer func() {
		fmt.Printf("Cleaning up temporary audio file: %s\n", audioPath)
		os.Remove(audioPath)
	}()

	// 2. Transcribe audio using OpenAI Whisper
	transcription, err := transcribeAudio(openaiKey, audioPath)
	if err != nil {
		log.Fatalf("Error transcribing audio: %v", err)
	}
	fmt.Println("\n--- Transcription (OpenAI Whisper) ---")
	fmt.Println(transcription)
	fmt.Println("--------------------------------------")

	// 3. Generate title/description using Ollama
	title, description, err := generateTitleDescription(ollamaEndpoint, generateModel, transcription)
	if err != nil {
		log.Fatalf("Error generating title/description: %v", err)
	}

	// 4. Print results
	fmt.Println("\n--- Generated YouTube Info ---")
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Description: %s\n", description)
	fmt.Println("-----------------------------")
}

// extractAudio extracts audio from a video file using ffmpeg and saves it as a temporary WAV file.
func extractAudio(videoPath string) (string, error) {
	fmt.Println("Extracting audio...")

	tempFile, err := os.CreateTemp("", "audio-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp audio file: %w", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()

	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found in PATH. Please install ffmpeg")
	}

	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", tempFileName)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return tempFileName, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	fmt.Printf("Running ffmpeg command: %s\n", strings.Join(cmd.Args, " "))
	if err := cmd.Start(); err != nil {
		os.Remove(tempFileName)
		return "", fmt.Errorf("failed to start ffmpeg command: %w", err)
	}

	slurp, _ := io.ReadAll(stderr)

	err = cmd.Wait()

	if len(slurp) > 0 {
		fmt.Fprintf(os.Stderr, "\n--- ffmpeg stderr ---\n%s\n---------------------\n", string(slurp))
	}

	if err != nil {
		os.Remove(tempFileName)
		return "", fmt.Errorf("ffmpeg command failed: %w", err)
	}

	fileInfo, statErr := os.Stat(tempFileName)
	if statErr != nil {
		os.Remove(tempFileName)
		return "", fmt.Errorf("failed to get stats for temp audio file %s: %w", tempFileName, statErr)
	}

	if fileInfo.Size() == 0 {
		os.Remove(tempFileName)
		return "", fmt.Errorf("ffmpeg ran successfully but produced an empty audio file. Check video source or ffmpeg logs (stderr printed above)")
	}

	fmt.Printf("Audio extracted successfully to: %s (Size: %d bytes)\n", tempFileName, fileInfo.Size())
	return tempFileName, nil
}

// transcribeAudio sends the audio file to OpenAI for transcription.
func transcribeAudio(openaiKey, audioPath string) (string, error) {
	fmt.Println("Transcribing audio using OpenAI Whisper...")

	client := openai.NewClient(openaiKey)
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioPath,
	}

	fmt.Printf("Sending audio file %s to OpenAI Whisper...\n", audioPath)

	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai transcription request failed: %w", err)
	}

	fmt.Println("Transcription received from OpenAI.")

	if resp.Text == "" {
		fmt.Println("Warning: OpenAI Whisper returned an empty transcription.")
	}

	return resp.Text, nil
}

// generateTitleDescription sends the transcription to Ollama to generate a title and description.
func generateTitleDescription(ollamaEndpoint, modelName, transcription string) (title string, description string, err error) {
	fmt.Println("Generating title and description...")

	parsedURL, err := url.Parse(ollamaEndpoint)
	if err != nil {
		err = fmt.Errorf("invalid ollama endpoint URL %q: %w", ollamaEndpoint, err)
		return
	}

	httpClient := &http.Client{
		Timeout: time.Minute * 5,
	}
	client := api.NewClient(parsedURL, httpClient)

	prompt := fmt.Sprintf("Based on the following video transcription, please generate a compelling YouTube title and description.\n\nTranscription:\n---\n%s\n---\n\nPlease format your response exactly like this:\nTitle: [Your generated title here]\nDescription: [Your generated description here]", transcription)

	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: func(b bool) *bool { return &b }(false),
	}

	fmt.Printf("Sending transcription to model %s for generation...\n", modelName)

	ctx := context.Background()
	var generatedText string
	err = client.Generate(ctx, req, func(res api.GenerateResponse) error {
		generatedText += res.Response
		if res.Done {
			fmt.Println("Title/Description received from Ollama.")
		}
		return nil
	})

	if err != nil {
		err = fmt.Errorf("ollama generation request failed: %w", err)
		return
	}

	if generatedText == "" {
		err = fmt.Errorf("ollama returned empty text for title/description")
		return
	}

	lines := strings.Split(generatedText, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		} else if strings.HasPrefix(line, "Description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		}
	}

	if title == "" || description == "" {
		fmt.Println("Warning: Could not parse title/description from Ollama response. Outputting raw response.")
		fmt.Println("Raw response:", generatedText)
		if title == "" {
			title = "Failed to parse title"
		}
		if description == "" {
			description = generatedText
		}
	}

	return title, description, nil
}
