import CustomTextArea from "$/components/form/custom-text-area";
import Input from "$/components/form/input";
import { useRobinMutation } from "$/lib/hooks";
import { slugify } from "$/lib/utils";
import stores from "$/stores";
import { valibotResolver } from "@hookform/resolvers/valibot";
import * as Form from "@radix-ui/react-form";
import {
	Badge,
	Box,
	Button,
	Card,
	DropdownMenu,
	Flex,
	Heading,
	Separator,
	Text,
} from "@radix-ui/themes";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as v from "valibot";
import { useSnapshot } from "valtio";

export const Route = createFileRoute("/_workspace/workspace/new")({
	component: RouteComponent,
});

const newWorkspaceSchema = v.object({
	name: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Name must be at least 2 characters"),
		v.maxLength(64, "Name must be at most 64 characters"),
		v.regex(
			/^[a-zA-Z0-9-_ ]+$/,
			"Name must contain only letters, numbers, and spaces",
		),
	),
	slug: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Slug must be at least 2 characters"),
		v.maxLength(80, "Slug must be at most 80 characters"),
		v.regex(
			/^[a-zA-Z0-9-]*$/,
			"Slug must contain only letters, numbers, and hyphens",
		),
	),
	description: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(512, "Cannot be more than 512 characters"),
		),
	),
});

type NewWorkspaceSchema = v.InferOutput<typeof newWorkspaceSchema>;

function RouteComponent() {
	const navigate = useNavigate();

	const workspace = useSnapshot(stores.workspace);
	const auth = useSnapshot(stores.auth);

	const form = useForm<NewWorkspaceSchema>({
		resolver: valibotResolver(newWorkspaceSchema),
	});

	const logoutMutation = useRobinMutation("auth.sign-out", {
		onSuccess: () => {
			auth.clear();
			return navigate({ to: "/auth/sign-in" });
		},
	});

	const mutation = useRobinMutation("workspace.create", {
		mode: "async",
		onSuccess: (data) => {
			workspace.addWorkspace(data.workspace); // We need to do this so that the initial load is not blank
			navigate({
				to: "/$workspaceSlug",
				params: { workspaceSlug: data.workspace.slug },
			});
			toast.success("Workspace created successfully");
		},
		setFormError: form.setError,
		invalidates: ["me"],
	});

	function handleName() {
		const name = form.getValues("name");
		const slug = form.getValues("slug");

		if (!slug) {
			form.setValue("slug", slugify(name));
		} else if (!name) {
			form.setValue("slug", "");
		}
	}

	return (
		<Flex direction="column" width="100vw" height="100vh">
			<Flex width="100%" justify="end" px="6" py="4">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						<Button variant="ghost">
							<Flex direction="column" align="start" gap="1">
								<Text color="gray" size="1">
									Signed in as:
								</Text>
								<Text size="1">{auth.user?.email}</Text>
							</Flex>
						</Button>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content align="end">
						<DropdownMenu.Item
							onSelect={() => logoutMutation.call()}
							disabled={logoutMutation.isMutating}
							style={{ width: "120px" }}
						>
							Sign out
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Flex>

			<Flex
				width="100%"
				height="100%"
				minHeight="0"
				flexGrow="1"
				direction="column"
				align="center"
				justify="center"
			>
				<Card variant="ghost">
					<Box width={{ initial: "95vw", xs: "385px" }} p="3">
						<Flex direction="column">
							<Heading size="6" weight="bold">
								New Workspace
							</Heading>

							<Text size="2" color="gray" mt="2">
								Workspaces are where you can organize your collections and
								documents, invite team members, and more.
							</Text>

							<Form.Root onSubmit={form.handleSubmit((d) => mutation.call(d))}>
								<Flex direction="column" gap="5" mt="4">
									<Flex direction="column" gap="3">
										<Input
											register={form.register}
											name="name"
											label="Name"
											errors={form.formState.errors}
											required
											textFieldProps={{ autoFocus: true, onBlur: handleName }}
										/>

										<Input
											register={form.register}
											name="slug"
											label="Slug"
											errors={form.formState.errors}
											required
											LeftSideComponent={() => (
												<Badge
													color="gray"
													size="2"
													variant="soft"
													radius="small"
													mx="-1"
												>
													{window.location.origin?.replace(/https?:\/\//, "")}/
												</Badge>
											)}
										/>

										<CustomTextArea
											register={form.register}
											name="description"
											label="Description"
											placeholder="What is this workspace for?"
											errors={form.formState.errors}
										/>
									</Flex>

									<Button type="submit" size="2" loading={mutation.isMutating}>
										Create Workspace
									</Button>
								</Flex>
							</Form.Root>
						</Flex>

						<Separator style={{ width: "100%" }} my="5" />

						<Text>
							Already invited? Click the link in the email to join the
							workspace.
						</Text>
					</Box>
				</Card>
			</Flex>
		</Flex>
	);
}
