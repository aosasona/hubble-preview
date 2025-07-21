import * as Form from "@radix-ui/react-form";
import { Flex, Text, TextArea, type TextAreaProps } from "@radix-ui/themes";
import FieldError from "./field-error";
import type {
	FieldErrors,
	FieldValues,
	Path,
	RegisterOptions,
	UseFormRegister,
} from "react-hook-form";
import { useMemo } from "react";

type RowTextAreaProps<T extends FieldValues = FieldValues> = {
	// Customisation
	label: string;
	name: Path<T>;
	placeholder?: string;
	textAreaProps?: TextAreaProps & React.RefAttributes<HTMLTextAreaElement>;

	// Form props
	register: UseFormRegister<T>;
	errors?: FieldErrors<T>;
	registerOptions?: RegisterOptions<T, Path<T>>;

	// Flags
	required?: boolean;
};
export default function RowTextArea<T extends FieldValues = FieldValues>(
	props: RowTextAreaProps<T>,
) {
	const registerOptions = useMemo(() => {
		const opts = {
			required: props.required ? "This field is required" : undefined,
		};

		if (props.registerOptions) {
			return {
				...opts,
				...props.registerOptions,
			};
		}

		return opts;
	}, [props.required, props.registerOptions]);

	return (
		<Form.Field name={props.name} className="w-full">
			<Flex
				direction={{ initial: "column", md: "row" }}
				align="start"
				justify={{ initial: "start", md: "between" }}
			>
				<Form.Label htmlFor={props.name}>
					<Text size="2" weight="medium" color="gray" highContrast>
						{props.label}
					</Text>
				</Form.Label>

				<Flex
					direction="column"
					minWidth={{ md: "320px" }}
					width={{ initial: "100%", md: "auto" }}
					align="start"
					mt="1"
					gap="1"
				>
					<TextArea
						{...props.register(props.name, registerOptions)}
						placeholder={props.placeholder}
						style={{ width: "100%" }}
						{...props.textAreaProps}
					/>

					<FieldError name={props.name} errors={props.errors} />
				</Flex>
			</Flex>
		</Form.Field>
	);
}
