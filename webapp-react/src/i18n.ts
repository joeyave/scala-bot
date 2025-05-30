import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import ru from "./locales/ru/translation.json";
import uk from "./locales/uk/translation.json";

i18n.use(initReactI18next);

export async function initI18n(lang: string) {
  await i18n.init({
    resources: {
      ru: { translation: ru },
      uk: { translation: uk },
    },
    lng: lang,
    fallbackLng: "uk",
    interpolation: {
      escapeValue: false,
    },
  });
}

export default i18n;
