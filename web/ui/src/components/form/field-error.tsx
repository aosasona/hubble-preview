import { Dot } from "@phosphor-icons/react";
import { Flex, Text } from "@radix-ui/themes";
import type { FieldErrors, FieldValues, Path } from "react-hook-form";

type FieldErrorProps<T extends FieldValues = FieldValues> = {
	errors?: FieldErrors<T>;
	name: Path<T>;
};

export default function FieldError<T extends FieldValues = FieldValues>(
	props: FieldErrorProps<T>,
) {
	if (!props.errors) return null;
	if (!props.errors?.[props.name]) return null;

	const errors = props.errors[props.name] as
		| FieldErrors<T>[Path<T>]
		| Array<FieldErrors<T>[Path<T>]>;

	return (
		<Flex mt="1">
			{Array.isArray(errors) ? (
				<Flex direction="column">
					{errors?.map((error, idx) => (
						<Flex key={`${error?.type}-${idx}`} align="center" ml="-1">
							<Dot size={16} color="var(--accent-12)" />
							<Text color="red" size="1">
								{error?.message?.toString() ?? ""}
							</Text>
						</Flex>
					))}
				</Flex>
			) : (
				<Text color="red" size="1">
					{errors?.message?.toString() ?? ""}
				</Text>
			)}
		</Flex>
	);
}
