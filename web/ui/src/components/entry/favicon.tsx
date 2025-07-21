import { Globe } from "@phosphor-icons/react";
import { Flex } from "@radix-ui/themes";
import { type JSX, useState } from "react";

type Props = {
	url: string;
	fallback?: () => JSX.Element;
};

// Renders a favicon image or a fallback component if the favicon fails to load
export default function Favicon(props: Props) {
	const [faviconError, setFaviconError] = useState(false);
	const hasFavicon = props.url && !faviconError;

	return (
		<Flex
			align="center"
			justify="center"
			className={`aspect-square size-5 rounded-md ${hasFavicon ? "border border-[var(--gray-5)]" : ""}`}
		>
			{hasFavicon ? (
				<img
					src={props.url}
					alt="favicon"
					className="h-full w-full rounded-md object-cover"
					onError={() => setFaviconError(true)}
				/>
			) : props.fallback ? (
				<props.fallback />
			) : (
				<Globe className="text-[var(--grass-11)]" size={20} />
			)}
		</Flex>
	);
}
