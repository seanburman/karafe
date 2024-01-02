import { PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";
import { useTheme } from "../context/theme";

export default function ScreenContainer ({ children }: PropsWithChildren){
    const theme = useTheme()
    return (
        <View
            style={[
                styles.container,
                {
                    backgroundColor: theme.colors.background
                },
            ]}
        >
            {children}
        </View>
    );
};

const styles = StyleSheet.create({
    container: {
        flex: 1,
        alignItems: "center",
        justifyContent: "center",
    },
});
