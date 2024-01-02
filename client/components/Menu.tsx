import {
    DrawerContentComponentProps,
    DrawerContentScrollView,
} from "@react-navigation/drawer";
import { Image, StyleSheet, Text, View } from "react-native";

export default function Menu(props: DrawerContentComponentProps) {
    return (
        <View style={styles.container}>
            <View style={styles.logoWrapper}>
                <Image
                    source={require("../assets/logo_transparent.png")}
                    style={styles.logo}
                />
                <Text style={styles.logoText}>Kache Krow</Text>
            </View>
            <DrawerContentScrollView style={styles.container}>
                <Text> Menu</Text>
            </DrawerContentScrollView>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
    },
    logoWrapper: {
        flexDirection: 'row',
        alignItems: 'center',
    },
    logo: {
        width: 100,
        height: 100,
    },
    logoText: {
        fontFamily: 'Alata, sans-serif',
        fontSize: 36,
        fontWeight: '600',
        fontStyle: 'normal'
    }
});
