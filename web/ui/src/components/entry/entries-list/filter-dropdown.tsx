import { DropdownMenu, Flex, Text } from "@radix-ui/themes";
import type { Filter, Mapping } from "./filter";
import stores from "$/stores";
import { useSnapshot } from "valtio";

type Props<T extends Filter["type"], Key extends string> = {
	type: T;
	category: string;
	items: Record<Key, Mapping>;
};

export default function FilterDropdown<
	T extends Filter["type"],
	Key extends string,
>(props: Props<T, Key>) {
	const filters = useSnapshot(stores.entriesList.filters);

	return (
		<DropdownMenu.Sub>
			<DropdownMenu.SubTrigger>{props.category}</DropdownMenu.SubTrigger>
			<DropdownMenu.SubContent>
				{Object.entries<Mapping>(props.items).map(([key, type]) => (
					<DropdownMenu.CheckboxItem
						key={key}
						checked={filters.some(
							(filter) =>
								filter.type === props.type && filter.value === (key as unknown),
						)}
						onCheckedChange={() =>
							stores.entriesList.toggleFilter({
								type: props.type,
								value: key,
							} as Filter)
						}
					>
						<Flex align="center" gap="2">
							<type.icon size={18} />
							<Text size="2">{type.label}</Text>
						</Flex>
					</DropdownMenu.CheckboxItem>
				))}
			</DropdownMenu.SubContent>
		</DropdownMenu.Sub>
	);
}
