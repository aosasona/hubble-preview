import { calculateColorHash } from "$/stores/workspace";
import { Flex, Grid, Text, Theme } from "@radix-ui/themes";
import { useMemo } from "react";

const SIZE_MAP = {
	1: 0.25,
	2: 0.5,
	3: 0.75,
	4: 1.0,
	5: 1.25,
	6: 1.5,
	7: 1.75,
	8: 2.0,
	9: 2.25,
} as const;

type Props = {
	/* The size of the icon */
	size?: keyof typeof SIZE_MAP;

	/* The name of the collection */
	name: string;
};

export function CollectionIcon(props: Props) {
	const accentColor = useMemo(() => {
		return calculateColorHash(props.name ?? "");
	}, [props.name]);

	const size = useMemo(() => props.size ?? 5, [props.size]);
	const sizeInRem = useMemo(() => `${SIZE_MAP[size]}rem`, [size]);
	const fontSize = useMemo(() => {
		return size <= 5 ? "1" : size < 7 ? "2" : "3";
	}, [size]);

	return (
		<Theme accentColor={accentColor}>
			<Flex
				align="center"
				justify="center"
				className="aspect-square rounded-[var(--radius-2)] border border-[var(--accent-8)]/25 bg-[var(--accent-9)]/10"
				style={{ width: sizeInRem, height: sizeInRem }}
				overflow="hidden"
			>
				<Text size={fontSize} weight="bold" color={accentColor}>
					{props.name.charAt(0)?.toUpperCase()}
				</Text>
			</Flex>
		</Theme>
	);
}

export default function CollectionIndicator(props: Props) {
	const accentColor = useMemo(() => {
		return calculateColorHash(props.name ?? "");
	}, [props.name]);

	return (
		<Theme accentColor={accentColor}>
			<Flex align="center" gap="2">
				<CollectionIcon name={props.name} size={5} />

				<Grid>
					<Text size="2" weight="medium" className="truncate">
						{props.name}
					</Text>
				</Grid>
			</Flex>
		</Theme>
	);
}
