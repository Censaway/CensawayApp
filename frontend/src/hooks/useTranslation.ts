import { useAppStore } from "../store/appStore";
import { translations } from "../locales/translations";

export const useTranslation = () => {
  const { settings } = useAppStore();
  const lang = (settings.language as "en" | "ru") || "en";

  const t = (path: string) => {
    const keys = path.split(".");
    let current: any = translations[lang];

    for (const key of keys) {
      if (current[key] === undefined) {
        return path;
      }
      current = current[key];
    }
    return current;
  };

  return { t, lang };
};