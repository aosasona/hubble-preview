import lightLogo from "$/assets/light-logo-transparent.svg";
import darkLogo from "$/assets/dark-logo-transparent.svg";
import { Box } from "@radix-ui/themes";
import { useMemo } from "react";

const SIZES = {
	xs: "12px",
	sm: "16px",
	md: "24px",
	lg: "32px",
} as const;

const PADDING: Record<keyof typeof SIZES, string> = {
	xs: "6px",
	sm: "var(--space-2)",
	md: "var(--space-2)",
	lg: "var(--space-2)",
};

type Props = {
	variant?: "bordered" | "plain";
	style?: React.CSSProperties;
	size?: keyof typeof SIZES;
};

export default function Logo({
	variant = "bordered",
	size = "sm",
	style,
}: Props) {
	const styles = useMemo(() => {
		switch (variant) {
			case "bordered":
				return {
					background: "var(--gray-3)",
					border: "2px solid var(--gray-4)",
					padding: PADDING[size],
					borderRadius: "var(--space-1)",
				};
			case "plain":
				return {};
		}
	}, [variant, size]);

	return (
		<Box width="max-content" style={{ ...styles, ...style }}>
			<img
				src={lightLogo}
				alt="logo"
				className="hidden aspect-square dark:block"
				width={SIZES[size]}
			/>
			<img
				src={darkLogo}
				alt="logo"
				className="block aspect-square dark:hidden"
				width={SIZES[size]}
			/>
		</Box>
	);
}
