import {
  mainButton,
  themeParamsButtonColor,
  themeParamsHintColor,
} from "@telegram-apps/sdk-react";

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
    backgroundColor: enabled
      ? themeParamsButtonColor()
      : themeParamsHintColor(),
    isLoaderVisible: loader,
  });
}
