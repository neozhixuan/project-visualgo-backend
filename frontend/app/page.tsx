// import WebSocketComponent from "../components/WebSocketComponent";
import { Box, Typography } from "@mui/material";
import CurrencyBitcoinIcon from "@mui/icons-material/CurrencyBitcoin";
import dynamic from "next/dynamic";

const WebSocketComponent = dynamic(
  () => import("../components/WebSocketComponent"),
  { ssr: false } // This will only import the component on the client-side
);

export default function Home() {
  return (
    <Box
      marginInline={"10%"}
      marginBlock={"5%"}
      style={{
        display: "flex",
        justifyContent: "center",
        flexDirection: "column",
      }}
    >
      <Box display={"flex"} flexDirection={"row"}>
        <CurrencyBitcoinIcon />
        <Typography>BNBBTC</Typography>
      </Box>
      <WebSocketComponent />
    </Box>
  );
}
