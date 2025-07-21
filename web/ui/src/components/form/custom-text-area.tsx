import * as Form from "@radix-ui/react-form";
import { Box, Text, TextArea, type TextAreaProps } from "@radix-ui/themes";
import FieldError from "./field-error";
import type {
	FieldErrors,
	FieldValues,
	Path,
	RegisterOptions,
	UseFormRegister,
} from "react-hook-form";
import { useMemo } from "react";

type CustomTextAreaProps<T extends FieldValues = FieldValues> = {
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
	hideLabel?: boolean;
	required?: boolean;
};
export default function CustomTextArea<T extends FieldValues = FieldValues>(
	props: CustomTextAreaProps<T>,
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
			{!props.hideLabel ? (
				<Form.Label htmlFor={props.name}>
					<Text size="2" weight="medium" color="gray">
						{props.label}
					</Text>
				</Form.Label>
			) : null}

			<Box width="100%" mt={props.hideLabel ? "0" : "1"}>
				<TextArea
					{...props.register(props.name, registerOptions)}
					placeholder={props.placeholder}
					{...props.textAreaProps}
				/>
			</Box>

			<FieldError name={props.name} errors={props.errors} />
		</Form.Field>
	);
}
