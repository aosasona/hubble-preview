import stores from "$/stores";
import { Heading, Text, Link, Flex } from "@radix-ui/themes";
import type { JSX } from "react";
import type { ExtraProps } from "react-markdown";

type PropType<Key extends keyof JSX.IntrinsicElements> = {
	[K in keyof JSX.IntrinsicElements[Key]]: JSX.IntrinsicElements[Key][K];
} & ExtraProps;

export function H1(props: PropType<"h1">) {
	return (
		<Heading size="8" mt="5" mb="1">
			{props.children}
		</Heading>
	);
}

export function H2(props: PropType<"h2">) {
	return (
		<Heading size="7" mt="5" mb="1">
			{props.children}
		</Heading>
	);
}

export function H3(props: PropType<"h3">) {
	return (
		<Heading size="6" mt="4" mb="1">
			{props.children}
		</Heading>
	);
}

export function H4(props: PropType<"h4">) {
	return (
		<Heading size="5" mt="3" mb="1">
			{props.children}
		</Heading>
	);
}

export function H5(props: PropType<"h5">) {
	return (
		<Heading size="4" mt="3" mb="1">
			{props.children}
		</Heading>
	);
}

export function H6(props: PropType<"h6">) {
	return (
		<Heading size="3" mt="3" mb="1">
			{props.children}
		</Heading>
	);
}

export function A(props: PropType<"a">) {
	return (
		<Link
			href={props.href}
			target="_blank"
			rel="noopener noreferrer"
			underline="always"
			color={stores.app.accentColor}
		>
			{props.children}
		</Link>
	);
}

export function P(props: PropType<"p">) {
	return (
		<Text
			as="p"
			size="2"
			mb="3"
			color="gray"
			style={{ lineHeight: 1.6 }}
			highContrast
		>
			{props.children}
		</Text>
	);
}

export function Blockquote(props: PropType<"blockquote">) {
	return <blockquote>{props.children}</blockquote>;
}

export function Code(props: PropType<"code">) {
	return (
		<code className="whitespace-pre-line" style={{ fontSize: "0.9em" }}>
			{props.children}
		</code>
	);
}

export function Img(props: PropType<"img">) {
	return (
		<Flex direction="column" align="center" gap="2" my="2">
			<img
				src={props.src}
				alt={props.alt}
				style={{
					maxWidth: "100%",
					maxHeight: "650px",
					objectFit: "contain",
				}}
			/>
			{props.alt?.trim() ? (
				<Text size="1" color="gray" align="center" style={{ maxWidth: "80%" }}>
					{props.alt}
				</Text>
			) : null}
		</Flex>
	);
}

export function Ul(props: PropType<"ul">) {
	return (
		<ul className="!my-2 !ml-4 !space-y-1 list-inside list-disc">
			{props.children}
		</ul>
	);
}

export function Ol(props: PropType<"ol">) {
	return (
		<ol className="!my-3 !ml-4 !space-y-1.5 list-inside list-decimal">
			{props.children}
		</ol>
	);
}

export function Li(props: PropType<"li">) {
	return <li>{props.children}</li>;
}
