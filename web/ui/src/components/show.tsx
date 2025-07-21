import { Fragment, type ReactNode } from "react";

type Props = {
	when: boolean;
	component?: ReactNode;
	children?: ReactNode;
};

export default function Show(props: Props) {
	if (!props.when) return null;

	if (!props.children && !props.component) {
		throw new Error(
			"Show component requires either a `component` or `children` prop.",
		);
	}

	if (props.component) {
		return <Fragment>{props.component}</Fragment>;
	}

	return <Fragment>{props.children}</Fragment>;
}
