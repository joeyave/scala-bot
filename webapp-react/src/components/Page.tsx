import { backButton } from "@tma.js/sdk-react";
import { type PropsWithChildren, useEffect } from "react";
import { useNavigate } from "react-router";

export function Page({
  children,
  back = true,
}: PropsWithChildren<{
  /**
   * True if it is allowed to go back from this page.
   */
  back?: boolean;
}>) {
  const navigate = useNavigate();

  useEffect(() => {
    if (back) {
      backButton.show();
      return backButton.onClick(() => {
        void navigate(-1);
      });
    }
    backButton.hide();
  }, [back, navigate]);

  return <>{children}</>;
}
