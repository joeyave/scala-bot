import SetlistSongDisplay, {
  SetlistSong,
} from "@/components/Setlist/SetlistSong.tsx";
import { Song } from "@/pages/EventPage/util/types.ts";
import { Section } from "@telegram-apps/telegram-ui";
import { useTranslation } from "react-i18next";
import {
  closestCenter,
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { hapticFeedback } from "@tma.js/sdk-react";
import {
  arrayMove,
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useState } from "react"; // const [, e] = bem("sortable-list");

// const [, e] = bem("sortable-list");

interface SetlistProps {
  items: Song[];
  onRemove: (song: Song) => void;
  onReorder: (newSetlist: Song[]) => void;
}

export function Setlist({ items, onRemove, onReorder }: SetlistProps) {
  const { t } = useTranslation();

  const [activeSong, setActiveSong] = useState<Song | null>(null);

  const handleDragStart = (event: DragStartEvent) => {
    hapticFeedback.impactOccurred("light");
    const { active } = event;
    setActiveSong(items.find((song) => song.id === active.id) ?? null);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (active.id !== over?.id) {
      const oldIndex = items.findIndex((s) => s.id === active.id);
      const newIndex = items.findIndex((s) => s.id === over?.id);
      const newSetlist = arrayMove(items, oldIndex, newIndex);
      onReorder(newSetlist);
    }
    setActiveSong(null);
  };

  const sensors = useSensors(
    // For mouse (desktop): start drag only after moving a few pixels
    useSensor(MouseSensor, {
      activationConstraint: {
        distance: 1,
      },
    }),

    // For touch (mobile): start drag only after short press delay
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 150,
        tolerance: 5,
      },
    }),
  );

  return (
    <div
      style={{
        userSelect: "none",
        WebkitUserSelect: "none",
        WebkitTouchCallout: "none",
        WebkitTapHighlightColor: "transparent",
      }}
    >
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragOver={() => {
          hapticFeedback.selectionChanged();
        }}

        // onDragPending={(event) => {
        //   console.log("pending", event);
        // }}
        // onDragMove={(event) => {
        //   console.log("move", event);
        // }}
        // onDragCancel={(event) => {
        //   console.log("cancel", event);
        // }}
        // onDragAbort={(event) => {
        //   console.log("abort", event);
        // }}
      >
        <SortableContext items={items} strategy={verticalListSortingStrategy}>
          <Section footer={items.length > 0 && t("setlistFooter")}>
            {items.map((song: Song) => (
              <SetlistSong key={song.id} song={song} onRemove={onRemove} />
            ))}
          </Section>
        </SortableContext>
        <DragOverlay>
          {activeSong ? (
            <SetlistSongDisplay song={activeSong} isOverlay />
          ) : null}
        </DragOverlay>
      </DndContext>
    </div>
  );
}
