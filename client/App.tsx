import "react-native-gesture-handler";
import React, { StyleSheet, View, useWindowDimensions } from "react-native";
import { NavigationContainer } from "@react-navigation/native";
import { createDrawerNavigator } from "@react-navigation/drawer";
import Home from "./screens/Home";
import Menu from "./components/Menu";
import { ThemeProvider } from "./context/theme";

const Drawer = createDrawerNavigator();
export default function App() {
    const { width } = useWindowDimensions();
    return (
        <ThemeProvider>
            <View style={styles.container}>
                <NavigationContainer>
                    <Drawer.Navigator
                        drawerContent={(props) => <Menu {...props} />}
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
            </View>
        </ThemeProvider>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
    },
});
