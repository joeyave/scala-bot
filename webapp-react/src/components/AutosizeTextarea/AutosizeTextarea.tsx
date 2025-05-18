import { Textarea, TextareaProps } from "@telegram-apps/telegram-ui";
import React, { useEffect, useRef } from "react";

interface AutosizeTextareaProps extends TextareaProps {
  maxHeight?: number | string;
}

export const AutosizeTextarea: React.FC<AutosizeTextareaProps> = ({
  maxHeight,
  style,
  ...props
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const textarea = containerRef.current?.querySelector("textarea");
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
      textarea.style.resize = "none";
      textarea.style.overflow = "hidden";
      if (maxHeight) {
        textarea.style.maxHeight =
          typeof maxHeight === "number" ? `${maxHeight}px` : maxHeight;
      }
    }
  }, [props.value, maxHeight]);

  return (
    <div ref={containerRef}>
      <Textarea
        {...props}
        style={{
          ...style,
          width: "100%",
          boxSizing: "border-box",
          resize: "none",
          overflow: "hidden",
          transition: "height 0.1s",
        }}
      />
    </div>
  );
};
