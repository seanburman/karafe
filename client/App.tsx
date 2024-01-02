import "react-native-gesture-handler";
import React, { StyleSheet, useWindowDimensions } from "react-native";
import { NavigationContainer } from "@react-navigation/native";
import { createDrawerNavigator } from "@react-navigation/drawer";
import Home from "./screens/Home";

const Drawer = createDrawerNavigator();

export default function App() {
  const { width } = useWindowDimensions();
    return (
        <NavigationContainer>
            <Drawer.Navigator
                initialRouteName="Home"
                screenOptions={{
                    // TODO: make these dimensions constant with theme hook
                    drawerType: width < 750 ? "front" : "permanent",
                    drawerPosition: "right",
                    overlayColor: "rgba(0,0,0,0.1)",
                    headerShown: false,
                    drawerStyle: {
                        borderColor: "#000000",
                        borderLeftWidth: 1,
                        paddingTop: 0,
                        marginTop: -4,
                    },
                }}
            >
                <Drawer.Screen name="Home" component={Home} />
            </Drawer.Navigator>
        </NavigationContainer>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: "#fff",
        alignItems: "center",
        justifyContent: "center",
    },
});
