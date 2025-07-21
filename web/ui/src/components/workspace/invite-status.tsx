import type { AccentColor } from "$/stores/app";
import type { Icon } from "@phosphor-icons/react";
import { Tooltip } from "@radix-ui/themes";

type InviteStatusProps = {
	icon: Icon;
	text: string;
	color: AccentColor;
};

export function InviteStatus(props: InviteStatusProps) {
	return (
		<Tooltip content={props.text}>
			<props.icon color={`var(--${props.color}-10)`} />
		</Tooltip>
	);
}
