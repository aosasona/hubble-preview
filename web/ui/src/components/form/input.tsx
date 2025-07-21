import { Flex, Text, TextField } from "@radix-ui/themes";
import type {
	FieldErrors,
	FieldValues,
	Path,
	RegisterOptions,
	UseFormRegister,
} from "react-hook-form";
import * as Form from "@radix-ui/react-form";
import { type FC, useMemo } from "react";
import FieldError from "./field-error";

type InputProps<T extends FieldValues = FieldValues> = {
	RightSideComponent?: FC;
	LeftSideComponent?: FC;
	// tslint:disable-next-line: no-any
	register: UseFormRegister<T>;
	label: string;
	name: Path<T>;
	type?:
		| "number"
		| "search"
		| "time"
		| "text"
		| "hidden"
		| "tel"
		| "url"
		| "email"
		| "date"
		| "datetime-local"
		| "month"
		| "password"
		| "week"
		| undefined;
	customField?: FC;
	textFieldProps?: TextField.RootProps & React.RefAttributes<HTMLInputElement>;
	registerOptions?: RegisterOptions<T, Path<T>>;
	errors?: FieldErrors<T>;
	hideLabel?: boolean;
	required?: boolean;
};

export default function Input<T extends FieldValues = FieldValues>(
	props: InputProps<T>,
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
			<Form.Control asChild>
				<Flex
					width="100%"
					align="center"
					mt={props.hideLabel ? "0" : "1"}
					gap="1"
				>
					{props.customField ? (
						<props.customField />
					) : (
						<TextField.Root
							{...props.register(props.name, registerOptions)}
							type={props.type}
							style={{ width: "100%" }}
							{...props.textFieldProps}
						>
							{props.RightSideComponent ? (
								<TextField.Slot side="right">
									<props.RightSideComponent />
								</TextField.Slot>
							) : null}

							{props.LeftSideComponent ? (
								<TextField.Slot side="left">
									<props.LeftSideComponent />
								</TextField.Slot>
							) : null}
						</TextField.Root>
					)}
				</Flex>
			</Form.Control>

			<FieldError errors={props.errors} name={props.name} />
		</Form.Field>
	);
}
