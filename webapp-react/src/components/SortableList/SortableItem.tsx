import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { bem } from "@/helpers/css/bem.ts";

import "./SortableList.css";
import { Icon16Cancel } from "@telegram-apps/telegram-ui/dist/icons/16/cancel";
import { logger } from "@/helpers/logger.ts";

const [, e] = bem("sortable-list");

export function SortableItem({ id, onRemove }) {
  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id: id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      className={e("item")}
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
    >
      <span data-song-id={id}>
        soooong {id}
      </span>
      <Icon16Cancel
        className={e("item-icon")}
        onClick={(e) => {
          logger.debug("click", { e });
          e.stopPropagation(); // prevent drag start
          onRemove?.(id);
        }}
      ></Icon16Cancel>
      {/*<i className={e("item-icon") + " fas fa-trash-alt song-remove"}></i>*/}
    </div>
  );
}
