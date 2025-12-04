use serde::{Deserialize, Serialize};
use std::process::Command;

#[derive(Debug, Serialize, Deserialize)]
struct AudioTrack {
    index: i32,
    language: String,
    title: String,
    codec: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct VideoInfo {
    #[serde(rename = "filePath")]
    file_path: String,
    #[serde(rename = "audioTracks")]
    audio_tracks: Vec<AudioTrack>,
}

#[derive(Debug, Serialize, Deserialize)]
struct GoResponse {
    success: bool,
    message: Option<String>,
    data: Option<serde_json::Value>,
}

// Get the path to the Go backend executable
fn get_go_backend_path() -> String {
    let exe_dir = std::env::current_exe()
        .ok()
        .and_then(|p| p.parent().map(|p| p.to_path_buf()))
        .unwrap_or_default();

    #[cfg(target_os = "windows")]
    let backend_name = "audio-track-backend.exe";
    #[cfg(not(target_os = "windows"))]
    let backend_name = "audio-track-backend";

    exe_dir.join(backend_name).to_string_lossy().to_string()
}

#[tauri::command]
async fn get_audio_tracks(video_path: String) -> Result<VideoInfo, String> {
    let backend_path = get_go_backend_path();

    let output = Command::new(&backend_path)
        .arg("get-tracks")
        .arg(&video_path)
        .output()
        .map_err(|e| format!("Failed to execute Go backend: {}", e))?;

    let stdout = String::from_utf8_lossy(&output.stdout);

    let response: GoResponse =
        serde_json::from_str(&stdout).map_err(|e| format!("Failed to parse response: {}", e))?;

    if !response.success {
        return Err(response
            .message
            .unwrap_or_else(|| "Unknown error".to_string()));
    }

    let video_info: VideoInfo = serde_json::from_value(response.data.unwrap())
        .map_err(|e| format!("Failed to parse video info: {}", e))?;

    Ok(video_info)
}

#[tauri::command]
async fn switch_audio_track(
    input_path: String,
    track_index: i32,
    output_path: String,
) -> Result<String, String> {
    let backend_path = get_go_backend_path();

    let output = Command::new(&backend_path)
        .arg("switch-track")
        .arg(&input_path)
        .arg(track_index.to_string())
        .arg(&output_path)
        .output()
        .map_err(|e| format!("Failed to execute Go backend: {}", e))?;

    let stdout = String::from_utf8_lossy(&output.stdout);

    let response: GoResponse =
        serde_json::from_str(&stdout).map_err(|e| format!("Failed to parse response: {}", e))?;

    if !response.success {
        return Err(response
            .message
            .unwrap_or_else(|| "Unknown error".to_string()));
    }

    Ok(response.message.unwrap_or_else(|| "Success".to_string()))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            get_audio_tracks,
            switch_audio_track
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
