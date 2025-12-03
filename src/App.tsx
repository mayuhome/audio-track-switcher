import { useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { open } from "@tauri-apps/plugin-dialog";
import "./App.css";

interface AudioTrack {
  index: number;
  language: string;
  title: string;
  codec: string;
}

interface VideoInfo {
  filePath: string;
  audioTracks: AudioTrack[];
}

function App() {
  const [videoPath, setVideoPath] = useState<string>("");
  const [videoInfo, setVideoInfo] = useState<VideoInfo | null>(null);
  const [selectedTrack, setSelectedTrack] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<string>("");
  const [processing, setProcessing] = useState<boolean>(false);

  async function selectVideoFile() {
    try {
      const selected = await open({
        multiple: false,
        filters: [
          {
            name: "Video",
            extensions: ["mp4", "mkv", "avi", "mov", "flv", "wmv", "webm"],
          },
        ],
      });

      if (selected && typeof selected === "string") {
        setVideoPath(selected);
        setMessage("");
        await loadAudioTracks(selected);
      }
    } catch (error) {
      setMessage(`Error selecting file: ${error}`);
    }
  }

  async function loadAudioTracks(path: string) {
    setLoading(true);
    setMessage("");
    try {
      const info = await invoke<VideoInfo>("get_audio_tracks", {
        videoPath: path,
      });
      setVideoInfo(info);
      if (info.audioTracks.length > 0) {
        setSelectedTrack(info.audioTracks[0].index);
      }
      setMessage(`Found ${info.audioTracks.length} audio track(s)`);
    } catch (error) {
      setMessage(`Error loading audio tracks: ${error}`);
      setVideoInfo(null);
    } finally {
      setLoading(false);
    }
  }

  async function switchAudioTrack() {
    if (!videoPath || !videoInfo) {
      setMessage("Please select a video file first");
      return;
    }

    setProcessing(true);
    setMessage("Processing...");

    try {
      // Generate output path
      const pathParts = videoPath.split(/[/\\]/);
      const fileName = pathParts[pathParts.length - 1];
      const fileNameWithoutExt = fileName.substring(0, fileName.lastIndexOf("."));
      const ext = fileName.substring(fileName.lastIndexOf("."));
      const outputPath = videoPath.replace(
        fileName,
        `${fileNameWithoutExt}_track${selectedTrack}${ext}`
      );

      await invoke("switch_audio_track", {
        inputPath: videoPath,
        trackIndex: selectedTrack,
        outputPath: outputPath,
      });

      setMessage(`Success! Output saved to: ${outputPath}`);
    } catch (error) {
      setMessage(`Error switching audio track: ${error}`);
    } finally {
      setProcessing(false);
    }
  }

  return (
    <div className="container">
      <header>
        <h1>🎵 Audio Track Switcher</h1>
        <p className="subtitle">Switch default audio track in video files</p>
      </header>

      <main>
        <section className="file-section">
          <button
            onClick={selectVideoFile}
            disabled={loading || processing}
            className="select-button"
          >
            📁 Select Video File
          </button>
          {videoPath && (
            <div className="file-info">
              <span className="label">Selected:</span>
              <span className="path">{videoPath}</span>
            </div>
          )}
        </section>

        {loading && (
          <div className="loading">
            <div className="spinner"></div>
            <p>Loading audio tracks...</p>
          </div>
        )}

        {videoInfo && !loading && (
          <section className="tracks-section">
            <h2>Audio Tracks ({videoInfo.audioTracks.length})</h2>
            <div className="tracks-list">
              {videoInfo.audioTracks.map((track) => (
                <div
                  key={track.index}
                  className={`track-item ${
                    selectedTrack === track.index ? "selected" : ""
                  }`}
                  onClick={() => setSelectedTrack(track.index)}
                >
                  <div className="track-header">
                    <input
                      type="radio"
                      name="track"
                      checked={selectedTrack === track.index}
                      onChange={() => setSelectedTrack(track.index)}
                    />
                    <span className="track-index">Track {track.index}</span>
                  </div>
                  <div className="track-details">
                    {track.language && (
                      <span className="track-lang">🌐 {track.language}</span>
                    )}
                    {track.title && (
                      <span className="track-title">📝 {track.title}</span>
                    )}
                    <span className="track-codec">🎧 {track.codec}</span>
                  </div>
                </div>
              ))}
            </div>

            <button
              onClick={switchAudioTrack}
              disabled={processing}
              className="process-button"
            >
              {processing ? "⏳ Processing..." : "✨ Switch Default Track"}
            </button>
          </section>
        )}

        {message && (
          <div className={`message ${message.includes("Error") ? "error" : "success"}`}>
            {message}
          </div>
        )}
      </main>

      <footer>
        <p>Powered by FFmpeg • Tauri • React • Go</p>
      </footer>
    </div>
  );
}

export default App;
