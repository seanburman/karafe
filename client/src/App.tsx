import { createDrawerNavigator } from "@react-navigation/drawer";
import { Box } from "native-base";
import React from "react-native";
import { NavigationContainer, DefaultTheme } from '@react-navigation/native';
import { Text, View } from "react-native";

const Drawer = createDrawerNavigator();

const MyTheme = {
    ...DefaultTheme,
    colors: {
      ...DefaultTheme.colors,
      primary: 'rgb(255, 45, 85)',
    },
  };

export default function App() {
    return (
        <NavigationContainer theme={MyTheme}>
            <Drawer.Navigator>
                <Drawer.Screen name="home" component={Home} />
            </Drawer.Navigator>
        </NavigationContainer>
    );
}

const Home: React.FC = () => {
    return (
        <View>
            <Text>
                Home
            </Text>
        </View>
    )
}
