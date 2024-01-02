import { AppRegistry } from "react-native";
import App from "./App";
import { NativeBaseProvider, Box } from "native-base";

export default function Main() {
    return (
        <NativeBaseProvider>
            <App/>
        </NativeBaseProvider>
    )
}

AppRegistry.registerComponent("App", () => Main);
AppRegistry.runApplication("App", { rootTag: document.getElementById("root") });
