import { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { open } from '@tauri-apps/plugin-dialog';
import { listen } from '@tauri-apps/api/event';
import './App.css';
import { useTranslation } from 'react-i18next';
import i18next from 'i18next';

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
  const [videoPath, setVideoPath] = useState<string>('');
  const [videoInfo, setVideoInfo] = useState<VideoInfo | null>(null);
  const [selectedTrack, setSelectedTrack] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<string>('');
  const [processing, setProcessing] = useState<boolean>(false);
  const [progress, setProgress] = useState<number>(0);
  const [language, setLanguage] = useState<string>('zh');
  const { t } = useTranslation();

  // Handle language change
  const handleLanguageChange = (newLanguage: string) => {
    i18next.changeLanguage(newLanguage);
    setLanguage(newLanguage);
  };

  async function selectVideoFile() {
    try {
      const selected = await open({
        multiple: false,
        filters: [
          {
            name: 'Video',
            extensions: ['mp4', 'mkv', 'avi', 'mov', 'flv', 'wmv', 'webm']
          }
        ]
      });

      if (selected && typeof selected === 'string') {
        setVideoPath(selected);
        setMessage('');
        await loadAudioTracks(selected);
      }
    } catch (error) {
      setMessage(`${t('app.error')}: ${t('app.errorSelectingFile')} ${error}`);
    }
  }

  useEffect(() => {
    // Listen for progress updates from Tauri backend
    const unlisten = listen('progress-update', (event) => {
      const progress = event.payload as number;
      setProgress(progress);
    });

    return () => {
      unlisten.then((fn) => fn());
    };
  }, []);

  async function loadAudioTracks(path: string) {
    setLoading(true);
    setMessage('');
    try {
      const info = await invoke<VideoInfo>('get_audio_tracks', {
        videoPath: path
      });
      setVideoInfo(info);
      if (info.audioTracks.length > 0) {
        setSelectedTrack(info.audioTracks[0].index);
      }
      setMessage(t('app.success', { count: info.audioTracks.length }));
    } catch (error) {
      setMessage(`${t('app.errorLoadingTracks')} ${error}`);
      setVideoInfo(null);
    } finally {
      setLoading(false);
    }
  }

  async function switchAudioTrack() {
    if (!videoPath || !videoInfo) {
      setMessage(t('app.pleaseSelectFile'));
      return;
    }

    setProcessing(true);
    setMessage(t('app.processing'));
    setProgress(0);

    try {
      // Get selected track info
      const selectedTrackInfo = videoInfo.audioTracks.find((track) => track.index === selectedTrack);
      const trackLanguage = selectedTrackInfo?.language || 'unknown';

      // Generate output path with track language
      const pathParts = videoPath.split(/[/\\]/);
      const fileName = pathParts[pathParts.length - 1];
      const fileNameWithoutExt = fileName.substring(0, fileName.lastIndexOf('.'));
      const ext = fileName.substring(fileName.lastIndexOf('.'));
      const outputPath = videoPath.replace(
        fileName,
        `${fileNameWithoutExt}_track${selectedTrack}_${trackLanguage}${ext}`
      );

      await invoke('switch_audio_track', {
        inputPath: videoPath,
        trackIndex: selectedTrack,
        outputPath: outputPath
      });

      setMessage(`${t('app.success')} ${t('app.outputSavedTo')}: ${outputPath}`);
    } catch (error) {
      setMessage(`${t('app.error')}: ${t('app.errorSwitchingTrack')} ${error}`);
    } finally {
      setProcessing(false);
      setProgress(0);
    }
  }

  return (
    <div className="container">
      <header>
        <div className="header-content">
          <div>
            <h1>{t('app.title')}</h1>
            <p className="subtitle">{t('app.subtitle')}</p>
          </div>
          <div className="language-switcher">
            <button
              className={`language-button ${language === 'zh' ? 'active' : ''}`}
              onClick={() => handleLanguageChange('zh')}
            >
              中文
            </button>
            <button
              className={`language-button ${language === 'en' ? 'active' : ''}`}
              onClick={() => handleLanguageChange('en')}
            >
              English
            </button>
          </div>
        </div>
      </header>

      <main>
        <section className="file-section">
          <button onClick={selectVideoFile} disabled={loading || processing} className="select-button">
            {t('app.selectButton')}
          </button>
          {videoPath && (
            <div className="file-info">
              <span className="label">{t('app.selected')}</span>
              <span className="path">{videoPath}</span>
            </div>
          )}
        </section>

        {loading && (
          <div className="loading">
            <div className="spinner"></div>
            <p>{t('app.loading')}</p>
          </div>
        )}

        {videoInfo && !loading && (
          <section className="tracks-section">
            <h2>{t('app.audioTracks', { count: videoInfo.audioTracks.length })}</h2>
            <div className="tracks-list">
              {videoInfo.audioTracks.map((track) => (
                <div
                  key={track.index}
                  className={`track-item ${selectedTrack === track.index ? 'selected' : ''}`}
                  onClick={() => setSelectedTrack(track.index)}
                >
                  <div className="track-header">
                    <input
                      type="radio"
                      name="track"
                      checked={selectedTrack === track.index}
                      onChange={() => setSelectedTrack(track.index)}
                    />
                    <span className="track-index">{t('app.track', { index: track.index })}</span>
                  </div>
                  <div className="track-details">
                    {track.language && (
                      <span className="track-lang">{t('app.language', { language: track.language })}</span>
                    )}
                    {track.title && <span className="track-title">{t('app.titleLabel', { title: track.title })}</span>}
                    <span className="track-codec">{t('app.codec', { codec: track.codec })}</span>
                  </div>
                </div>
              ))}
            </div>

            <button onClick={switchAudioTrack} disabled={processing} className="process-button">
              {processing ? t('app.processing') : t('app.switchButton')}
            </button>
          </section>
        )}

        {processing && (
          <div className="processing-container">
            <div className="progress-container">
              <div className="progress-bar" style={{ width: `${progress}%` }}></div>
              <span className="progress-text">{Math.round(progress)}%</span>
            </div>
          </div>
        )}

        {message && <div className={`message ${message.includes('Error') ? 'error' : 'success'}`}>{message}</div>}
      </main>

      <footer>
        <p>{t('app.footer')}</p>
      </footer>
    </div>
  );
}

export default App;
