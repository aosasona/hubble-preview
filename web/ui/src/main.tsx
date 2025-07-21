import "@radix-ui/themes/styles.css";
import "./index.css";

import ReactDOM from "react-dom/client";
import App from "./app";

// Render the app
const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
	const root = ReactDOM.createRoot(rootElement);
	root.render(<App />);
}
