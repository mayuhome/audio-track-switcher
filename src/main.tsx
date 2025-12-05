import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { I18nextProvider } from 'react-i18next';
import i18next from 'i18next';
import zh from './locales/zh.json';
import en from './locales/en.json';

// Configure i18next for the app
i18next.init({
  resources: {
    zh: {
      translation: zh
    },
    en: {
      translation: en
    }
  },
  lng: 'zh', // Set default language to Chinese
  fallbackLng: 'zh',
  interpolation: {
    escapeValue: false
  },
  react: {
    useSuspense: false
  }
});

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <I18nextProvider i18n={i18next}>
      <App />
    </I18nextProvider>
  </React.StrictMode>
);
