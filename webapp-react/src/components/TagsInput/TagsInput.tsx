import { useTranslation } from "react-i18next";
import { arraysEqualMultiset } from "@/helpers/multiselect.ts";
import { Multiselect } from "@telegram-apps/telegram-ui";
import { MultiselectOption } from "@telegram-apps/telegram-ui/dist/components/Form/Multiselect/types";

export function TagsInput({
  options,
  value,
  onChange,
  placeholder,
  creatable,
}: {
  options: MultiselectOption[];
  value: MultiselectOption[];
  onChange?: (v: MultiselectOption[]) => void;
  placeholder?: string;
  creatable?: string | boolean;
}) {
  const { t } = useTranslation();

  // Provide default values using translation if not supplied
  placeholder = placeholder ?? t("tagPlaceholder");
  creatable = creatable ?? t("tagCreatable");
  const handleChange = (newOptions: MultiselectOption[]) => {
    // Removing unselected items if exactly the same items were selected again.
    const areEqual = arraysEqualMultiset(value, newOptions, (a, b) => {
      return a.value === b.value;
    });

    if (areEqual) {
      newOptions.pop();
    }

    if (onChange) {
      onChange(newOptions);
    }
  };

  return (
    <Multiselect
      options={options}
      value={value}
      onChange={handleChange}
      placeholder={placeholder}
      creatable={creatable}
    />
  );
}
