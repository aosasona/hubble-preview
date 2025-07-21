export type OneOf<T> = T extends (infer U)[] ? U : T;

export type Nullable<T extends Record<string, unknown>> = {
	[P in keyof T]: T[P] | null;
};

// e.g FieldOf<{ name: string, meta: { val: string } }, "name"> is string
// e.g FieldOf<{ name: string, meta: { val: string } }, "meta.val"> is string
export type FieldOf<T, K extends string> = K extends `${infer L}.${infer R}`
	? L extends keyof T
		? FieldOf<T[L], R>
		: never
	: K extends keyof T
		? T[K]
		: never;

// Make all fields in T readonly recursively
export type ReadonlyAll<T> = {
	readonly [P in keyof T]: T[P] extends Record<string, unknown>
		? ReadonlyAll<T[P]>
		: T[P];
};

export const MEMBER_ROLE = {
	guest: "Guest",
	user: "Member",
	admin: "Admin",
	owner: "Owner",
} as const;

export type MemberRole = keyof typeof MEMBER_ROLE;
