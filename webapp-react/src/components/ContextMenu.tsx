import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/react";
import { FC, ReactNode } from "react";

export interface ContextMenuItem {
  label: string;
  onClick: () => void;
  destructive?: boolean;
}

interface ContextMenuProps {
  items: ContextMenuItem[];
  as?: any;
  trigger?: ReactNode;
  children?: ReactNode;
  [key: string]: any;
}

export const ContextMenu: FC<ContextMenuProps> = ({
  items,
  as = "div",
  trigger,
  children,
  ...props
}) => {
  return (
    <Menu>
      {({ open }) => (
        <>
          {/* Backdrop overlay for mobile focus and dimming, rendered only when open */}
          {open && (
            <div className="fixed inset-0 z-[9998] bg-black/[0.04] dark:bg-black/[0.15] transition-opacity duration-100" />
          )}

          {/* Polymorphic trigger button that propagates props and stops event bubbling */}
          <MenuButton
            as={as}
            onPointerDown={(e: React.PointerEvent) => {
              e.stopPropagation();
            }}
            onMouseDown={(e: React.MouseEvent) => {
              e.stopPropagation();
            }}
            onClick={(e: React.MouseEvent) => {
              e.stopPropagation();
            }}
            {...props}
          >
            {trigger || children}
          </MenuButton>

          {/* Floating Popover Menu anchored to the button, focus ring disabled */}
          <MenuItems
            transition
            anchor="bottom end"
            className="z-[9999] min-w-[200px] rounded-xl border border-black/[0.08] dark:border-white/[0.08] bg-[var(--tg-theme-bg-color,#ffffff)] p-1 shadow-lg backdrop-blur-xl transition duration-100 ease-out data-[closed]:scale-95 data-[closed]:opacity-0 focus:outline-none"
          >
            <div className="flex flex-col">
              {items.map((item, idx) => (
                <MenuItem key={idx}>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      item.onClick();
                    }}
                    className={`w-full text-left px-4 py-2.5 text-sm font-medium rounded-lg transition-colors duration-75 active:bg-black/[0.05] dark:active:bg-white/[0.05] ${
                      item.destructive
                        ? "text-[var(--tg-theme-destructive-text-color,#ff3b30)]"
                        : "text-[var(--tg-theme-text-color,#000000)]"
                    }`}
                  >
                    {item.label}
                  </button>
                </MenuItem>
              ))}
            </div>
          </MenuItems>
        </>
      )}
    </Menu>
  );
};
