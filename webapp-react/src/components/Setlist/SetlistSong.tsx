import { Song } from "@/pages/EventPage/util/types.ts";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { TrashIcon } from "@heroicons/react/20/solid";
import { IconButton } from "@telegram-apps/telegram-ui";
import React from "react";
import "./Setlist.css";
import { hapticFeedback } from "@tma.js/sdk-react";

interface SetlistSongDisplayProps extends React.HTMLAttributes<HTMLDivElement> {
  song: Song;
  onRemove?: (song: Song) => void;
  isOverlay?: boolean;
}

const SetlistSongDisplay = React.forwardRef<
  HTMLDivElement,
  SetlistSongDisplayProps
>(({ song, onRemove, isOverlay, ...props }, ref) => {
  return (
    <div
      className={`relative mb-1 flex items-center justify-between rounded-xl bg-[var(--tg-theme-section-bg-color)] px-4 py-3 ${isOverlay ? "z-10 cursor-grabbing shadow-lg" : "cursor-grab"} `}
      ref={ref}
      {...props}
    >
      <div className="mr-2 min-w-0 flex-1">
        <div className="truncate font-[family-name:var(--tgui--font-family)] text-[length:var(--tgui--text--font_size)] leading-[var(--tgui--text--line_height)] font-[var(--tgui--font_weight--accent3)] text-[var(--tg-theme-text-color,#000)]">
          {song.name}
        </div>
        <div className="truncate font-[family-name:var(--tgui--font-family)] text-[length:var(--tgui--subheadline2--font_size)] leading-[var(--tgui--subheadline2--line_height)] font-[var(--tgui--font_weight--accent3)] text-[var(--tg-theme-hint-color)]">
          {`${song.key || "?"}, ${song.bpm || "?"}, ${song.time || "?"}`}
        </div>
      </div>

      <IconButton
        // className="![background:color-mix(in_srgb,var(--tg-theme-destructive-text-color)_20%,transparent)]"
        onClick={(e) => {
          hapticFeedback.impactOccurred("light");
          e.stopPropagation();
          onRemove?.(song);
        }}
        mode="outline"
        size="s"
      >
        <TrashIcon className="h-5 w-5 text-[var(--tg-theme-destructive-text-color)]" />
      </IconButton>
    </div>
  );
});

export default SetlistSongDisplay;

SetlistSongDisplay.displayName = "SetlistSongDisplay";

interface SetlistSongProps {
  song: Song;
  onRemove: (id: Song) => void;
}

export function SetlistSong({ song, onRemove }: SetlistSongProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: song.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0 : 1,
  };

  return (
    <SetlistSongDisplay
      ref={setNodeRef}
      style={style}
      song={song}
      onRemove={onRemove}
      {...attributes}
      {...listeners}
    />
  );
}
