import { mainButton, themeParams } from "@tma.js/sdk-react";

export function setMainButton({
  visible,
  text,
  enabled,
  loader,
}: {
  visible?: boolean;
  text?: string;
  enabled?: boolean;
  loader?: boolean;
}) {
  mainButton.setParams({
    text: text,
    isVisible: visible,
    isEnabled: enabled,
    bgColor: enabled
      ? themeParams.buttonColor()
      : themeParams.hintColor(),
    isLoaderVisible: loader,
  });
}
