package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// AudioTrack represents an audio track in a video file
type AudioTrack struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Codec    string `json:"codec"`
}

// VideoInfo represents video file information
type VideoInfo struct {
	FilePath    string       `json:"filePath"`
	AudioTracks []AudioTrack `json:"audioTracks"`
}

// Response represents the JSON response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GetAudioTracks retrieves all audio tracks from a video file using ffprobe
func GetAudioTracks(videoPath string) (*VideoInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-select_streams", "a",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ffprobe: %w", err)
	}

	var result struct {
		Streams []struct {
			Index int `json:"index"`
			Tags  struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
			CodecName string `json:"codec_name"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	videoInfo := &VideoInfo{
		FilePath:    videoPath,
		AudioTracks: make([]AudioTrack, 0),
	}

	for _, stream := range result.Streams {
		track := AudioTrack{
			Index:    stream.Index,
			Language: stream.Tags.Language,
			Title:    stream.Tags.Title,
			Codec:    stream.CodecName,
		}
		videoInfo.AudioTracks = append(videoInfo.AudioTracks, track)
	}

	return videoInfo, nil
}

// SwitchDefaultAudioTrack changes the default audio track of a video file
func SwitchDefaultAudioTrack(inputPath string, trackIndex int, outputPath string) error {
	// Build ffmpeg command to set the default audio track
	// We'll copy all streams but set the disposition of the selected audio track as default
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-map", "0",
		"-c", "copy",
		"-disposition:a", "0", // Remove default flag from all audio tracks
		fmt.Sprintf("-disposition:a:%d", trackIndex), "default", // Set selected track as default
		"-y", // Overwrite output file
		"-progress", "pipe:1", // Output progress to stdout
		outputPath,
	)

	// Get stdout pipe to read progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Create a scanner to read stdout line by line
	scanner := bufio.NewScanner(stdout)
	var progressMap = make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Empty line means end of progress update
			if progressMap["out_time_ms"] != "" && progressMap["duration"] != "" {
				// Parse progress
				outTimeMs, _ := strconv.ParseFloat(progressMap["out_time_ms"], 64)
				duration, _ := strconv.ParseFloat(progressMap["duration"], 64)
				if duration > 0 {
					percent := (outTimeMs / duration) * 100
					// Output progress as JSON
					progressResponse := Response{
						Success: true,
						Message: "progress",
						Data: map[string]interface{}{
							"progress": percent,
						},
					}
					output, _ := json.Marshal(progressResponse)
					fmt.Println(string(output))
				}
			}
			// Reset progress map
			progressMap = make(map[string]string)
		} else {
			// Parse key-value pair
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				progressMap[parts[0]] = parts[1]
			}
		}
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		// Capture stderr for error information
		stderr, _ := cmd.StderrPipe()
		stderrOutput, _ := io.ReadAll(stderr)
		return fmt.Errorf("ffmpeg error: %w\nOutput: %s", err, string(stderrOutput))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ffmpeg output: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printError("No command specified")
		return
	}

	command := os.Args[1]

	switch command {
	case "get-tracks":
		if len(os.Args) < 3 {
			printError("Video path not specified")
			return
		}
		videoPath := os.Args[2]
		
		info, err := GetAudioTracks(videoPath)
		if err != nil {
			printError(err.Error())
			return
		}
		
		printSuccess("Audio tracks retrieved successfully", info)

	case "switch-track":
		if len(os.Args) < 5 {
			printError("Usage: switch-track <input_path> <track_index> <output_path>")
			return
		}
		
		inputPath := os.Args[2]
		trackIndex := 0
		fmt.Sscanf(os.Args[3], "%d", &trackIndex)
		outputPath := os.Args[4]

		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create output directory: %v", err))
			return
		}

		err := SwitchDefaultAudioTrack(inputPath, trackIndex, outputPath)
		if err != nil {
			printError(err.Error())
			return
		}
		
		printSuccess("Audio track switched successfully", map[string]interface{}{
			"outputPath": outputPath,
		})

	default:
		printError(fmt.Sprintf("Unknown command: %s", command))
	}
}

func printSuccess(message string, data interface{}) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	output, _ := json.Marshal(response)
	fmt.Println(string(output))
}

func printError(message string) {
	response := Response{
		Success: false,
		Message: message,
	}
	output, _ := json.Marshal(response)
	fmt.Println(string(output))
}
