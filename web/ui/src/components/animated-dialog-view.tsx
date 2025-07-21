import type { ReactNode } from "@tanstack/react-router";
import { motion } from "motion/react";
import Show from "./show";

type Props = {
	children: ReactNode;
	key: string | number;
	visible: boolean;
};

const animation = {
	initial: { x: 100 },
	animate: { x: 0 },
	exit: { x: -100 },
	transition: { type: "tween", duration: 0.125 },
};

export default function AnimatedDialogView(props: Props) {
	return (
		<Show when={props.visible}>
			<motion.div key={props.key} {...animation}>
				{props.children}
			</motion.div>
		</Show>
	);
}
