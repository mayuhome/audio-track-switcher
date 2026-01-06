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
func SwitchDefaultAudioTrack(inputPath string, trackIndex int, audioTrackPosition int, outputPath string) error {
	// Build ffmpeg command to set the default audio track
	// We'll copy all streams but set the disposition of the selected audio track as default
	args := []string{
		"-i", inputPath,
		"-map", "0",
		"-c", "copy",
		"-y",                  // Overwrite output file
		"-progress", "pipe:1", // Output progress to stdout
	}

	// Set all audio tracks to not be default
	args = append(args, "-disposition:a", "none")
	// Set selected track as default
	args = append(args, fmt.Sprintf("-disposition:a:%d", audioTrackPosition), "default")
	// Add output path
	args = append(args, outputPath)

	ffmpegCmd := exec.Command("ffmpeg", args...)

	// Get stdout pipe to read progress
	stdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := ffmpegCmd.Start(); err != nil {
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
	if err := ffmpegCmd.Wait(); err != nil {
		// Capture stderr for error information
		stderr, _ := ffmpegCmd.StderrPipe()
		stderrOutput, _ := io.ReadAll(stderr)
		return fmt.Errorf("ffmpeg error: %w\nOutput: %s", err, string(stderrOutput))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ffmpeg output: %w", err)
	}

	return nil
}

// RemoveAudioTrack removes an audio track from a video file
func RemoveAudioTrack(inputPath string, trackIndex int, outputPath string) error {
	// Get audio tracks to verify there's more than one track
	videoInfo, err := GetAudioTracks(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get audio tracks: %w", err)
	}

	if len(videoInfo.AudioTracks) <= 1 {
		return fmt.Errorf("cannot remove the only audio track")
	}

	// Find the selected track in the audio streams
	var selectedAudioTrack *AudioTrack
	for _, track := range videoInfo.AudioTracks {
		if track.Index == trackIndex {
			selectedAudioTrack = &track
			break
		}
	}

	if selectedAudioTrack == nil {
		return fmt.Errorf("Track index %d not found in audio tracks", trackIndex)
	}

	// Build ffmpeg command to remove the selected audio track
	// We'll explicitly map all streams except the selected audio track
	args := []string{
		"-i", inputPath,
		"-c", "copy", // Copy streams without re-encoding
		"-y",                  // Overwrite output file
		"-progress", "pipe:1", // Output progress to stdout
	}

	// Get all streams from input file using ffprobe
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute ffprobe: %w", err)
	}

	var result struct {
		Streams []struct {
			Index int    `json:"index"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Map all streams except the selected audio track
	for _, stream := range result.Streams {
		if stream.CodecType == "audio" && stream.Index == selectedAudioTrack.Index {
			// Skip the selected audio track
			continue
		}
		// Map all other streams
		args = append(args, "-map", fmt.Sprintf("0:%d", stream.Index))
	}

	// Add output path
	args = append(args, outputPath)

	ffmpegCmd := exec.Command("ffmpeg", args...)

	// Get stdout pipe to read progress
	stdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Start the command
	if err := ffmpegCmd.Start(); err != nil {
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
	if err := ffmpegCmd.Wait(); err != nil {
		// Capture stderr for error information
		stderr, _ := ffmpegCmd.StderrPipe()
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
		outputPath := GenerateUniqueFilePath(os.Args[4])

		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create output directory: %v", err))
			return
		}

		// Get audio tracks to find the relative position
		videoInfo, err := GetAudioTracks(inputPath)
		if err != nil {
			printError(fmt.Sprintf("Failed to get audio tracks: %v", err))
			return
		}

		// Find the relative position of the selected track in the audio streams
		audioTrackPosition := -1
		for i, track := range videoInfo.AudioTracks {
			if track.Index == trackIndex {
				audioTrackPosition = i
				break
			}
		}

		if audioTrackPosition == -1 {
			printError(fmt.Sprintf("Track index %d not found in audio tracks", trackIndex))
			return
		}

		err = SwitchDefaultAudioTrack(inputPath, trackIndex, audioTrackPosition, outputPath)
		if err != nil {
			printError(err.Error())
			return
		}

		printSuccess("Audio track switched successfully", map[string]interface{}{
			"outputPath": outputPath,
		})

	case "remove-track":
		if len(os.Args) < 5 {
			printError("Usage: remove-track <input_path> <track_index> <output_path>")
			return
		}

		inputPath := os.Args[2]
		trackIndex := 0
		fmt.Sscanf(os.Args[3], "%d", &trackIndex)
		outputPath := GenerateUniqueFilePath(os.Args[4])

		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create output directory: %v", err))
			return
		}

		err := RemoveAudioTrack(inputPath, trackIndex, outputPath)
		if err != nil {
			printError(err.Error())
			return
		}

		printSuccess("Audio track removed successfully", map[string]interface{}{
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

// GenerateUniqueFilePath checks if a file exists and generates a unique file path with incrementing number
func GenerateUniqueFilePath(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
	ext := filepath.Ext(filename)

	for i := 1; i < 1000; i++ {
		newFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		newFilePath := filepath.Join(dir, newFileName)
		if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
			return newFilePath
		}
	}

	return filePath
}

func printError(message string) {
	response := Response{
		Success: false,
		Message: message,
	}
	output, _ := json.Marshal(response)
	fmt.Println(string(output))
}
