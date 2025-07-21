import { toTitleCase } from "$/lib/utils";
import stores from "$/stores";
import type { Parent, HeaderItem } from "$/stores/layout";
import layoutStore from "$/stores/layout";
import { Container, Flex, Heading } from "@radix-ui/themes";
import { AnimatePresence, motion } from "motion/react";
import { type JSX, type ReactNode, memo, useEffect, useMemo } from "react";
import { useSnapshot } from "valtio";

type HeadingProp = string | { title: string; component: () => JSX.Element };

type Props = {
	/* The title of the page */
	title?: string;

	/* The text or component to render at the top of the page */
	heading: HeadingProp;

	/* Whether to show the header in the body */
	showHeader?: boolean;

	/* If true, the layout will take up the full screen and will not use container boundaries */
	fullScreen?: boolean;

	/* The header items to display */
	header?: {
		/* The parent of the header items - this is usually the parent route, for example: "settings" */
		parent?: Parent;

		/* The header items to display alongside the parent, if any */
		items?: HeaderItem[];

		/* Whether to retain the parent on navigate (i.e. when the underlying route changes) */
		retainParentOnNavigate?: boolean;

		/* Whether to use the parent in the title */
		hideParentInTitle?: boolean;
	};

	/* The children to render */
	children: ReactNode;
};

function RawHeader(props: { content: HeadingProp }) {
	return (
		<Heading size="8">
			{typeof props.content === "string" ? props.content : props.content.title}
		</Heading>
	);
}

const Header = memo(RawHeader);

export default function PageLayout(props: Props) {
	const layout = useSnapshot(stores.layout);

	// biome-ignore lint/correctness/useExhaustiveDependencies: I don't want to re-run this effect when the layout store changes because it's not necessary and would cause an infinite re-render
	useEffect(() => {
		layoutStore.fullScreen = props.fullScreen;

		if (props?.header?.parent) {
			layout.setParent(props.header.parent);
		}

		// Use the header items if they are provided
		if (props?.header?.items?.length) {
			// If we already have a parent, append the items to the parent
			if (props.header.parent) {
				layout.appendItems(props.header.items);
				return;
			}

			layout.setHeaderItems(props?.header?.items);
		}

		// Use the heading if no header items are provided and the heading is a string
		if (!props?.header?.items?.length) {
			layout.appendHeaderItem(
				typeof props.heading === "string"
					? { title: props.heading }
					: props.heading,
			);
		}

		return () => {
			if (props.header?.parent && props.header.retainParentOnNavigate) {
				return layout.clearChildrenItems();
			}

			layout.clearHeaderItems();
		};
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [props.header, props.heading, props.fullScreen]);

	const title = useMemo(() => {
		if (props.title) {
			return props.title;
		}

		if (layout.headerItems.length) {
			return layout.headerItems.map((item) => item.title).join(" > ");
		}

		if (typeof props.heading === "string") {
			if (props.header?.parent && !props.header.hideParentInTitle) {
				return `${toTitleCase(props.header.parent)} > ${toTitleCase(props.heading)}`;
			}

			return props.heading;
		}

		return props.heading.title;
	}, [
		props?.header?.parent,
		props.heading,
		props.title,
		props.header?.hideParentInTitle,
		layout.headerItems,
	]);

	return (
		<AnimatePresence>
			<motion.div
				initial={{ opacity: 0, y: 10 }}
				animate={{ opacity: 1, y: 0 }}
				exit={{ opacity: 0, y: 10 }}
				transition={{ duration: 0.1, type: "tween" }}
				className="flex h-full min-h-0 w-full flex-grow"
			>
				{props.fullScreen ? (
					<Flex
						direction="column"
						flexGrow="1"
						minHeight="0"
						width="100%"
						height="100%"
						overflow="hidden"
					>
						{props.showHeader ? <Header content={props.heading} /> : null}
						<title>{title}</title>

						{props.children}
					</Flex>
				) : (
					<Flex
						direction="column"
						flexGrow="1"
						minHeight="0"
						width="100%"
						height="100%"
					>
						<Container
							width="100%"
							height="100%"
							minHeight="0"
							size={{
								initial: "4",
								xs: "2",
								sm: "2",
								md: "2",
								lg: "4",
								xl: "2",
							}}
							align={{ initial: "center", md: "left" }}
							px={{ initial: "4", md: "6" }}
							py={{ initial: "4", md: "6" }}
						>
							{props.showHeader ? <Header content={props.heading} /> : null}
							<title>{title}</title>

							{props.children}
						</Container>
					</Flex>
				)}
			</motion.div>
		</AnimatePresence>
	);
}
