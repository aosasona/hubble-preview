import { defineConfig } from "vite";
import path from "node:path";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { TanStackRouterVite } from "@tanstack/router-plugin/vite";

// https://vite.dev/config/
export default defineConfig({
	plugins: [tailwindcss(), TanStackRouterVite(), react()],
	resolve: {
		alias: {
			$: path.resolve(__dirname, "src"),
			"@ui": path.resolve(__dirname, "src/components"),
			"@lib": path.resolve(__dirname, "src/lib"),
			"@stores": path.resolve(__dirname, "src/stores"),
			"@assets": path.resolve(__dirname, "src/assets"),
			"@images": path.resolve(__dirname, "src/assets/images"),
		},
	},
});
