import { X } from "@phosphor-icons/react";
import { Dialog, Flex, IconButton, Progress, Text } from "@radix-ui/themes";
import { AnimatePresence, motion } from "motion/react";

type ProgressProps = {
	/**
	 * @description The total number of steps
	 */
	total: number;

	/**
	 * @description The number of steps completed, this should start at 1 and go up to `total`
	 */
	completed: number;

	/**
	 * @description Whether the progress indicator is in a dialog and should show a close button
	 */
	inDialog?: boolean;
};

export default function ProgressIndicator(props: ProgressProps) {
	return (
		<AnimatePresence>
			<motion.div
				initial={{ opacity: 0, y: -10 }}
				animate={{ opacity: 1, y: 0 }}
				exit={{ opacity: 0, y: -10 }}
				transition={{ type: "tween", duration: 0.125 }}
			>
				<Flex align="start" justify="between" mb="5">
					<Flex direction="column" gap="2">
						<Flex gap="2" width="150px">
							{Array.from({ length: props.total }).map((_, i) => (
								<Progress
									key={`${i + 1}`}
									value={props.completed >= i + 1 ? 100 : 0}
								/>
							))}
						</Flex>

						<Text size="1" color="gray" weight="medium">
							STEP {props.completed} of {props.total}
						</Text>
					</Flex>

					{props.inDialog && props.completed < props.total ? (
						<Dialog.Close>
							<IconButton variant="ghost" color="gray">
								<X size={18} />
							</IconButton>
						</Dialog.Close>
					) : null}
				</Flex>
			</motion.div>
		</AnimatePresence>
	);
}
