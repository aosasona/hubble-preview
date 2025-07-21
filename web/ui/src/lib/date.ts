export function formatDateString(dateWithTz: string) {
	const date = new Date(dateWithTz);
	const formatedDate = date.toLocaleDateString("en-US", {
		year: "numeric",
		month: "long",
		day: "numeric",
	});
	const formatedTime = date.toLocaleTimeString("en-US", {
		hour: "2-digit",
		minute: "numeric",
		second: "numeric",
		hour12: true,
		timeZoneName: "short",
	});

	return `${formatedDate} ${formatedTime}`;
}

export function timeAgo(dateWithTz: string) {
	const date = new Date(dateWithTz);
	const now = new Date();
	const diff = now.getTime() - date.getTime();
	const seconds = Math.floor(diff / 1000);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);
	const days = Math.floor(hours / 24);
	const months = Math.floor(days / 30);

	if (months > 0) {
		return `${months} month${months > 1 ? "s" : ""} ago`;
	}

	if (days > 0) {
		return `${days} day${days > 1 ? "s" : ""} ago`;
	}

	if (hours > 0) {
		return `${hours} hour${hours > 1 ? "s" : ""} ago`;
	}

	if (minutes > 0) {
		return `${minutes} minute${minutes > 1 ? "s" : ""} ago`;
	}

	return `${seconds} second${seconds > 1 ? "s" : ""} ago`;
}
