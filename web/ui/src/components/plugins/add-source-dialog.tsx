import {
	Dialog,
	Button,
	Flex,
	IconButton,
	DataList,
	Skeleton,
	Heading,
	Card,
	Badge,
	Tooltip,
	Text,
} from "@radix-ui/themes";
import { useRobinMutation } from "$/lib/hooks";
import { useForm } from "react-hook-form";
import * as v from "valibot";
import * as Form from "@radix-ui/react-form";
import Input from "../form/input";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { MagnifyingGlass } from "@phosphor-icons/react";
import type { Workspace, MembershipStatus } from "$/lib/server/types";
import Show from "../show";
import { extractError } from "$/lib/error";
import { toast } from "sonner";

type Props = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	data: {
		workspace: Workspace;
		status: MembershipStatus;
	};
};

const lookupFormSchema = v.object({
	url: v.pipe(
		v.string(),
		v.nonEmpty("URL is required"),
		v.transform((url) => url.toLowerCase()),
	),
});

type LookupFormSchema = v.InferOutput<typeof lookupFormSchema>;

export default function AddSourceDialog(props: Props) {
	const lookupForm = useForm<LookupFormSchema>({
		resolver: valibotResolver(lookupFormSchema),
	});

	const sourceLookupMutation = useRobinMutation("plugin.source.find", {
		retry: 2,
		setFormError: lookupForm.setError,
	});

	const addSourceMutation = useRobinMutation("plugin.source.add", {
		onSuccess: (data) => {
			toast.success(
				`Added "${data.source.name}" source to "${props.data.workspace.name}" workspace`,
			);
			onOpenChange(false);
		},
		setFormError: lookupForm.setError,
		invalidates: ["plugin.source.list"],
		retry: false,
	});

	function onOpenChange(open: boolean) {
		if (!open) {
			lookupForm.reset();
		}
		props.onOpenChange(open);
		sourceLookupMutation.reset();
	}

	return (
		<Dialog.Root open={props.open} onOpenChange={onOpenChange}>
			<Dialog.Content maxWidth="450px">
				<Dialog.Title size="6" mb="2">
					Add a new source
				</Dialog.Title>
				<Dialog.Description size="2" color="gray">
					Sources are decentralized collections that can be used to organize and
					share plugins.
				</Dialog.Description>

				<Flex direction="column" width="100%" height="100%" gap="3" mt="3">
					<Form.Root
						onSubmit={lookupForm.handleSubmit((d) =>
							sourceLookupMutation.call({
								url: d.url,
								workspace_id: props.data.workspace.id,
							}),
						)}
					>
						<Flex gap="2" align="start">
							<Input
								register={lookupForm.register}
								name="url"
								label="URL"
								errors={lookupForm.formState.errors}
								required
								hideLabel
								textFieldProps={{
									placeholder: "https://github.com/keystroke-tools/hub",
								}}
							/>
							<IconButton
								type="submit"
								loading={lookupForm.formState.isSubmitting}
							>
								<MagnifyingGlass />
							</IconButton>
						</Flex>
					</Form.Root>

					<Show when={sourceLookupMutation.isError}>
						<Flex py="3" gap="2" align="center" justify="center">
							<Text color="red" align="center">
								{extractError(sourceLookupMutation.error)?.message}
							</Text>
						</Flex>
					</Show>

					<Show when={sourceLookupMutation.isSuccess}>
						<Skeleton loading={sourceLookupMutation.isMutating}>
							<Card>
								<Flex direction="column" gap="4">
									<Heading size="4" weight="regular" color="gray">
										Properties
									</Heading>

									<DataList.Root>
										<DataList.Item>
											<DataList.Label>Name</DataList.Label>
											<DataList.Value>
												{sourceLookupMutation.data?.source?.name ?? "unknown"}
											</DataList.Value>
										</DataList.Item>
										<DataList.Item>
											<DataList.Label>Author</DataList.Label>
											<DataList.Value>
												{sourceLookupMutation.data?.source?.author ?? "unknown"}
											</DataList.Value>
										</DataList.Item>
										<DataList.Item>
											<DataList.Label>URL</DataList.Label>
											<DataList.Value>
												{sourceLookupMutation.data?.source?.url ?? "unknown"}
											</DataList.Value>
										</DataList.Item>
										<DataList.Item>
											<DataList.Label>Description</DataList.Label>
											<DataList.Value>
												{sourceLookupMutation.data?.source?.description ??
													"unknown"}
											</DataList.Value>
										</DataList.Item>
										<DataList.Item>
											<DataList.Label>Versioning Strategy</DataList.Label>
											<DataList.Value>
												<Badge color="gray" variant="surface">
													{sourceLookupMutation.data?.source
														?.versioning_strategy ?? "unknown"}
												</Badge>
											</DataList.Value>
										</DataList.Item>
										<DataList.Item>
											<DataList.Label>Plugins</DataList.Label>
											<DataList.Value>
												<Flex wrap="wrap" gap="2">
													{sourceLookupMutation.data?.plugins.map((plugin) => (
														<Tooltip
															key={plugin.name}
															content={plugin.description}
														>
															<Badge color="gray" variant="surface">
																{plugin.name}
															</Badge>
														</Tooltip>
													))}
												</Flex>
											</DataList.Value>
										</DataList.Item>
									</DataList.Root>
								</Flex>
							</Card>
						</Skeleton>
					</Show>
				</Flex>

				<Flex
					direction="row-reverse"
					gap="3"
					align="center"
					justify="start"
					mt="4"
				>
					<Button
						type="submit"
						variant="solid"
						size="2"
						loading={addSourceMutation.isMutating}
						disabled={!sourceLookupMutation.data}
						onClick={() => {
							addSourceMutation.call({
								workspace_id: props.data.workspace.id,
								url: sourceLookupMutation.data?.source.url ?? "",
							});
						}}
					>
						Continue
					</Button>
					<Dialog.Close>
						<Button type="button" variant="soft" color="gray" size="2">
							Cancel
						</Button>
					</Dialog.Close>
				</Flex>
			</Dialog.Content>
		</Dialog.Root>
	);
}
