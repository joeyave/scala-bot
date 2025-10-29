import { bem } from "@/helpers/css/bem.ts";
import { SortableItem } from "@/components/SortableList/SortableItem.tsx";

const [, e] = bem("sortable-list");

export function SortableList({ items, onRemove }) {
  return (
    <div className={e("container")}>
      {items.map((id) => (
        <SortableItem key={id} id={id} onRemove={onRemove} />
      ))}
    </div>
  );
}
