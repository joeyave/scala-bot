import { classNames } from "@/helpers/css/classnames.ts";
import {
  Button as HeadlessButton,
  Checkbox as HeadlessCheckbox,
  Combobox,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
  Disclosure,
  DisclosureButton,
  DisclosurePanel,
  Input as HeadlessInput,
  Select as HeadlessSelect,
  Textarea as HeadlessTextarea,
} from "@headlessui/react";
import {
  CheckIcon,
  ChevronUpDownIcon,
  XMarkIcon,
} from "@heroicons/react/20/solid";
import {
  ComponentPropsWithoutRef,
  ElementType,
  HTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
  createContext,
  useContext,
  useMemo,
  useState,
} from "react";

import "./ui.css";

type Status = "error" | undefined;
type Size = "s" | "m" | "l";

export interface MultiselectOption {
  value: string;
  label: string;
}

export function AppRoot({
  appearance,
  platform,
  className,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  appearance?: "light" | "dark";
  platform?: "ios" | "base" | string;
}) {
  return (
    <div
      {...props}
      data-appearance={appearance}
      data-platform={platform}
      className={classNames("ui-app-root", className)}
    />
  );
}

export function List({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return <div {...props} className={classNames("ui-list", className)} />;
}

function SectionRoot({
  header,
  footer,
  className,
  children,
  ...props
}: HTMLAttributes<HTMLElement> & {
  header?: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <section {...props} className={classNames("ui-section", className)}>
      {header ? <div className="ui-section__header">{header}</div> : null}
      <div className="ui-section__body">{children}</div>
      {footer ? <div className="ui-section__footer">{footer}</div> : null}
    </section>
  );
}

export function SectionHeader({
  large,
  className,
  ...props
}: HTMLAttributes<HTMLDivElement> & { large?: boolean }) {
  return (
    <div
      {...props}
      className={classNames(
        "ui-section-header",
        large && "ui-section__header--large",
        className,
      )}
    />
  );
}

export const Section = Object.assign(SectionRoot, { Header: SectionHeader });

export function Cell({
  before,
  after,
  subtitle,
  subhead,
  readOnly,
  multiline,
  className,
  children,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  before?: ReactNode;
  after?: ReactNode;
  subtitle?: ReactNode;
  subhead?: ReactNode;
  readOnly?: boolean;
  multiline?: boolean;
}) {
  return (
    <div
      {...props}
      className={classNames(
        "ui-cell",
        readOnly && "ui-cell--readonly",
        multiline && "ui-cell--multiline",
        className,
      )}
    >
      {before ? <div className="ui-cell__before">{before}</div> : null}
      <div className="ui-cell__content">
        {subhead ? <div className="ui-cell__subhead">{subhead}</div> : null}
        <div className="ui-cell__title">{children}</div>
        {subtitle ? <div className="ui-cell__subtitle">{subtitle}</div> : null}
      </div>
      {after ? <div className="ui-cell__after">{after}</div> : null}
    </div>
  );
}

function ControlShell({
  header,
  status,
  after,
  className,
  children,
}: {
  header?: ReactNode;
  status?: Status;
  after?: ReactNode;
  className?: string;
  children: ReactNode;
}) {
  return (
    <label className="ui-control-shell">
      {header ? <span className="ui-control-shell__label">{header}</span> : null}
      <span
        className={classNames(
          "ui-control",
          status === "error" && "ui-control--error",
          className,
        )}
      >
        {children}
        {after ? <span className="ui-control__after">{after}</span> : null}
      </span>
    </label>
  );
}

export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  after?: ReactNode;
  header?: ReactNode;
  status?: Status;
};

export function Input({
  after,
  header,
  status,
  className,
  ...props
}: InputProps) {
  return (
    <ControlShell after={after} header={header} status={status}>
      <HeadlessInput
        {...props}
        className={classNames("ui-input", className)}
      />
    </ControlShell>
  );
}

export type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  after?: ReactNode;
  header?: ReactNode;
  status?: Status;
};

export function Textarea({
  after,
  header,
  status,
  className,
  ...props
}: TextareaProps) {
  return (
    <ControlShell after={after} header={header} status={status}>
      <HeadlessTextarea
        {...props}
        className={classNames("ui-textarea", className)}
      />
    </ControlShell>
  );
}

export type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  after?: ReactNode;
  header?: ReactNode;
  status?: Status;
};

export function Select({
  after,
  header,
  status,
  className,
  children,
  ...props
}: SelectProps) {
  return (
    <ControlShell after={after} header={header} status={status}>
      <HeadlessSelect {...props} className={classNames("ui-select", className)}>
        {children}
      </HeadlessSelect>
    </ControlShell>
  );
}

type ButtonOwnProps<C extends ElementType> = {
  Component?: C;
  mode?: "filled" | "bezeled" | "outline" | "plain" | "gray";
  size?: Size;
  stretched?: boolean;
  before?: ReactNode;
  after?: ReactNode;
  loading?: boolean;
  children?: ReactNode;
  className?: string;
};

export function Button<C extends ElementType = "button">({
  Component,
  mode = "filled",
  size = "m",
  stretched,
  before,
  after,
  loading,
  className,
  children,
  disabled,
  ...props
}: ButtonOwnProps<C> &
  Omit<ComponentPropsWithoutRef<C>, keyof ButtonOwnProps<C>>) {
  return (
    <HeadlessButton
      {...props}
      as={(Component ?? "button") as ElementType}
      disabled={disabled || loading}
      className={classNames(
        "ui-button",
        `ui-button--${mode}`,
        `ui-button--${size}`,
        stretched && "ui-button--stretched",
        className,
      )}
    >
      {loading ? <Spinner size="s" /> : before}
      {children}
      {after}
    </HeadlessButton>
  );
}

export function IconButton({
  mode = "plain",
  size = "m",
  className,
  type = "button",
  ...props
}: ComponentPropsWithoutRef<"button"> & {
  mode?: "plain" | "outline" | "bezeled";
  size?: Size;
}) {
  return (
    <button
      {...props}
      type={type}
      className={classNames(
        "ui-icon-button",
        `ui-icon-button--${mode}`,
        `ui-icon-button--${size}`,
        className,
      )}
    />
  );
}

export function Chip<C extends ElementType = "span">({
  Component,
  mode = "outline",
  className,
  children,
  ...props
}: {
  Component?: C;
  mode?: "mono" | "outline";
  className?: string;
  children?: ReactNode;
} & Omit<ComponentPropsWithoutRef<C>, "className" | "children">) {
  const Tag = Component ?? "span";
  return (
    <Tag
      {...props}
      className={classNames("ui-chip", `ui-chip--${mode}`, className)}
    >
      {children}
    </Tag>
  );
}

export function Spinner({ size = "m" }: { size?: Size }) {
  return <span className={classNames("ui-spinner", `ui-spinner--${size}`)} />;
}

export function Placeholder({
  header,
  description,
  className,
  children,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  header?: ReactNode;
  description?: ReactNode;
}) {
  return (
    <div {...props} className={classNames("ui-placeholder", className)}>
      {children}
      {header ? <div className="ui-placeholder__header">{header}</div> : null}
      {description ? (
        <div className="ui-placeholder__description">{description}</div>
      ) : null}
    </div>
  );
}

export function Checkbox({
  className,
  checked,
  disabled,
  onChange,
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <HeadlessCheckbox
      {...props}
      checked={checked}
      disabled={disabled}
      onChange={(nextChecked) => {
        if (!onChange) {
          return;
        }
        onChange({
          target: { checked: nextChecked },
          currentTarget: { checked: nextChecked },
        } as unknown as React.ChangeEvent<HTMLInputElement>);
      }}
      className={classNames("ui-check", className)}
    />
  );
}

export function Radio({
  className,
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      type="radio"
      className={classNames("ui-check", className)}
    />
  );
}

export function Title({
  level = "1",
  weight,
  className,
  ...props
}: HTMLAttributes<HTMLHeadingElement> & {
  level?: "1" | "2" | "3";
  weight?: "1" | "2" | "3";
}) {
  const Tag = `h${level}` as ElementType;
  return (
    <Tag
      {...props}
      data-weight={weight}
      className={classNames("ui-title", `ui-title--${level}`, className)}
    />
  );
}

export function Headline({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return <div {...props} className={classNames("ui-headline", className)} />;
}

export function Text({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div {...props} className={classNames("ui-text", className)} />;
}

export function Multiselect({
  options,
  value,
  onChange,
  placeholder,
  creatable,
}: {
  options: MultiselectOption[];
  value: MultiselectOption[];
  onChange?: (value: MultiselectOption[]) => void;
  placeholder?: string;
  creatable?: string | boolean;
}) {
  const [query, setQuery] = useState("");
  const normalizedQuery = query.trim().toLowerCase();
  const selectedValues = useMemo(
    () => new Set(value.map((option) => option.value)),
    [value],
  );
  const availableOptions = options.filter((option) => {
    if (selectedValues.has(option.value)) {
      return false;
    }
    if (!normalizedQuery) {
      return true;
    }
    return option.label.toLowerCase().includes(normalizedQuery);
  });
  const canCreate =
    !!creatable &&
    query.trim().length > 0 &&
    !selectedValues.has(query.trim()) &&
    !options.some((option) => option.value === query.trim());

  const removeOption = (option: MultiselectOption) => {
    onChange?.(value.filter((current) => current.value !== option.value));
  };

  const handleChange = (nextOptions: MultiselectOption[]) => {
    const seen = new Set<string>();
    const deduped = nextOptions.filter((option) => {
      if (!option.value || seen.has(option.value)) {
        return false;
      }
      seen.add(option.value);
      return true;
    });
    onChange?.(deduped);
    setQuery("");
  };

  return (
    <Combobox
      multiple
      immediate
      by="value"
      value={value}
      onChange={handleChange}
      onClose={() => setQuery("")}
    >
      <div className="ui-combobox">
        <div className="ui-combobox__control">
          {value.map((option) => (
            <button
              key={option.value}
              type="button"
              className="ui-combobox__tag"
              onClick={() => removeOption(option)}
            >
              <span>{option.label}</span>
              <XMarkIcon aria-hidden className="ui-combobox__tag-icon" />
            </button>
          ))}
          <ComboboxInput
            aria-label={placeholder}
            className="ui-combobox__input"
            value={query}
            placeholder={value.length === 0 ? placeholder : undefined}
            onChange={(event) => setQuery(event.target.value)}
          />
          <ChevronUpDownIcon aria-hidden className="ui-combobox__chevron" />
        </div>

        <ComboboxOptions
          anchor="bottom"
          className="ui-combobox__options empty:invisible"
        >
          {canCreate ? (
            <ComboboxOption
              value={{ value: query.trim(), label: query.trim() }}
              className="ui-combobox__option--creatable"
            >
              <CheckIcon aria-hidden className="ui-combobox__check" style={{ color: "transparent" }} />
              <span>
                {typeof creatable === "string" ? creatable : query.trim()}
              </span>
              {typeof creatable === "string" ? (
                <span className="ui-combobox__option-muted">
                  {query.trim()}
                </span>
              ) : null}
            </ComboboxOption>
          ) : null}

          {availableOptions.length === 0 && !canCreate ? (
            <div className="ui-combobox__empty">{placeholder}</div>
          ) : null}

        {availableOptions.map((option) => (
            <ComboboxOption
              key={option.value}
              value={option}
              className="ui-combobox__option"
            >
              <CheckIcon aria-hidden className="ui-combobox__check" />
              <span>{option.label}</span>
            </ComboboxOption>
          ))}
        </ComboboxOptions>
      </div>
    </Combobox>
  );
}

const AccordionContext = createContext<{
  expanded: boolean;
  onChange: (expanded: boolean) => void;
} | null>(null);

function AccordionRoot({
  expanded,
  onChange,
  className,
  ...props
}: Omit<HTMLAttributes<HTMLDivElement>, "onChange"> & {
  expanded: boolean;
  onChange: (expanded: boolean) => void;
}) {
  return (
    <AccordionContext.Provider value={{ expanded, onChange }}>
      <Disclosure
        {...props}
        as="div"
        key={String(expanded)}
        defaultOpen={expanded}
        data-expanded={expanded}
        className={classNames("ui-accordion", className)}
      />
    </AccordionContext.Provider>
  );
}

function AccordionSummary({
  className,
  multiline,
  ...props
}: ComponentPropsWithoutRef<"button"> & { multiline?: boolean }) {
  const context = useContext(AccordionContext);
  return (
    <DisclosureButton
      {...props}
      aria-expanded={context?.expanded}
      data-multiline={multiline}
      className={classNames("ui-accordion__summary", className)}
      onClick={(event) => {
        props.onClick?.(event);
        context?.onChange(!context.expanded);
      }}
    />
  );
}

function AccordionContent({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  const context = useContext(AccordionContext);
  if (!context?.expanded) {
    return null;
  }
  return (
    <DisclosurePanel
      {...props}
      static
      className={classNames("ui-accordion__content", className)}
    />
  );
}

export const Accordion = Object.assign(AccordionRoot, {
  Summary: AccordionSummary,
  Content: AccordionContent,
});
