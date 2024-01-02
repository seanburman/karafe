import React, { AppRegistry } from "react-native";
import App from "./App";
import "react-native-gesture-handler";

export default function Main() {
    return <App />;
}

AppRegistry.registerComponent("App", () => Main);
AppRegistry.runApplication("App", { rootTag: document.getElementById("root") });
