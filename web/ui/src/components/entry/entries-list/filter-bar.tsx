import stores from "$/stores";
import { Badge, Button, Flex, IconButton, Text } from "@radix-ui/themes";
import { useSnapshot } from "valtio";
import { motion } from "motion/react";
import { X } from "@phosphor-icons/react";
import { CATEGORY_MAPPING, ENTRY_TYPE_MAPPING, STATUS_MAPPING } from "./filter";
import { useCallback } from "react";

export default function FilterBar() {
	const workspace = useSnapshot(stores.workspace);
	const filters = useSnapshot(stores.entriesList.filters);

	const findCollectionName = useCallback(
		(collectionId: string) => {
			const collection = workspace.activeWorkspaceCollections.find(
				(c) => c.id === collectionId,
			);
			return collection ? collection.name : "";
		},
		[workspace.activeWorkspaceCollections],
	);

	return (
		<motion.div
			initial={{ opacity: 0, y: -30 }}
			animate={{ opacity: 1, y: 0 }}
			exit={{ opacity: 0, y: -30 }}
			transition={{ duration: 0.15 }}
			className="flex w-full max-w-full items-center justify-between border-[var(--gray-4)] border-b bg-background px-[var(--space-4)] py-[var(--space-2)]"
		>
			<Flex align="center" gap="2" wrap="wrap" width="100%">
				{filters.map((filter) => (
					<motion.div
						key={`${filter.type}-${filter.value}`}
						initial={{ opacity: 0, scale: 0.9 }}
						animate={{ opacity: 1, scale: 1 }}
						exit={{ opacity: 0, scale: 0.9 }}
						transition={{ duration: 0.175 }}
					>
						<Badge color="gray" variant="soft" size="2" radius="large">
							<Flex gap="3" align="center">
								<Flex gap="2" align="center">
									<Text size="1" color="gray" weight="medium">
										{CATEGORY_MAPPING[filter.type]} is{" "}
										<Text
											size="1"
											color="gray"
											weight="bold"
											as="span"
											highContrast
										>
											{filter.type === "type"
												? ENTRY_TYPE_MAPPING[filter.value].label
												: filter.type === "status"
													? STATUS_MAPPING[filter.value].label
													: filter.type === "collection"
														? findCollectionName(filter.value)
														: filter.value}
										</Text>
									</Text>
								</Flex>

								<IconButton
									variant="ghost"
									color="gray"
									size="1"
									onClick={() => stores.entriesList.removeFilter(filter)}
								>
									<X size={16} />
								</IconButton>
							</Flex>
						</Badge>
					</motion.div>
				))}
			</Flex>

			<Button
				variant="ghost"
				color="gray"
				size="1"
				onClick={() => stores.entriesList.clearFilters()}
			>
				<Flex align="center" gap="1">
					<Text size="1">Clear all</Text>
				</Flex>
			</Button>
		</motion.div>
	);
}
