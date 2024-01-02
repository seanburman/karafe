import { StyleSheet, Text, View } from "react-native";
import { useTheme } from "../context/theme";
import ScreenContainer from "../components/Container";

export default function Stores() {
    const theme = useTheme();

    return (
        <ScreenContainer>
            <Text>Home</Text>
        </ScreenContainer>
    );
}

const styles = StyleSheet.create({
    container: {
      
    },
});
