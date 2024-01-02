import { StyleSheet, Text, View } from "react-native";
import { useTheme } from "../context/theme";

export default function Home() {
    const theme = useTheme();

    return (
        <View
            style={[
                styles.container,
                {
                    backgroundColor: theme.colors.background,
                },
            ]}
        >
            <Text>Home</Text>
        </View>
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
