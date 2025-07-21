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
	required?: boolean;
};

export default function RowInput<T extends FieldValues = FieldValues>(
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
			<Flex
				direction={{ initial: "column", md: "row" }}
				align={{ initial: "start", md: "center" }}
				justify={{ initial: "start", md: "between" }}
			>
				<Form.Label htmlFor={props.name}>
					<Text size="2" weight="medium" color="gray" highContrast>
						{props.label}
					</Text>
				</Form.Label>
				<Form.Control asChild>
					<Flex
						direction="column"
						minWidth={{ md: "250px" }}
						width={{ initial: "100%", md: "auto" }}
						align="start"
						mt="1"
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
						<FieldError errors={props.errors} name={props.name} />
					</Flex>
				</Form.Control>
			</Flex>
		</Form.Field>
	);
}
