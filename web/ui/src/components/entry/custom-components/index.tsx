import type { ComponentType, JSX } from "react";
import type { ExtraProps } from "react-markdown";
import * as c from "./components";

type Components = {
	[Key in keyof JSX.IntrinsicElements]?:
		| ComponentType<JSX.IntrinsicElements[Key] & ExtraProps>
		| keyof JSX.IntrinsicElements;
};

export const CUSTOM_COMPONENTS: Components = {
	h1: c.H1,
	h2: c.H2,
	h3: c.H3,
	h4: c.H4,
	h5: c.H5,
	h6: c.H6,
	a: c.A,
	p: c.P,
	blockquote: c.Blockquote,
	img: c.Img,
	code: c.Code,
	ul: c.Ul,
	ol: c.Ol,
	li: c.Li,
};
