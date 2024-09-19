"use client";
import React, { useEffect, useState } from "react";
import ReactApexChart from "react-apexcharts";
import { ApexOptions } from "apexcharts";
import useMediaQuery from "@mui/material/useMediaQuery";
import { Box, SxProps, Theme } from "@mui/material";

interface KlineDataPoint {
  x: Date;
  y: number[];
}

interface SeriesData {
  data: KlineDataPoint[];
}

interface ScrollableBoxProps {
  children: React.ReactNode;
}

interface ScrollableBoxProps {
  children: React.ReactNode;
  customSx?: SxProps<Theme>; // Allows accepting additional styles via the `sx` prop
  minHeight?: string | number;
  width?: string | number;
  padding?: string | number;
  overflow?: "auto" | "hidden" | "scroll" | "visible";
  alignSelf?: "flex-start" | "flex-end" | "center" | "baseline" | "stretch";
}

const ScrollableBox: React.FC<ScrollableBoxProps> = ({
  children,
  minHeight,
  width,
  padding,
  overflow,
  alignSelf,
  customSx,
}) => {
  // Default styles with hidden scrollbar
  const defaultSx: SxProps<Theme> = {
    overflow: "scroll",
    "&::-webkit-scrollbar": {
      display: "none", // WebKit browsers
    },
    scrollbarWidth: "none", // Firefox
    minHeight: minHeight,
    width: width,
    padding: padding,
  };

  // Combine default styles with any custom styles passed in
  const mergedSx = { ...defaultSx, ...customSx };

  return <Box sx={mergedSx}>{children}</Box>;
};

const WebSocketComponent = () => {
  const [data, setData] = useState<any[]>([]);
  const [series, setSeries] = useState<SeriesData[]>([{ data: [] }]);
  const [ema, setEma] = useState<any[]>([]);
  const tablet = useMediaQuery("(min-width:960px)");
  const wssUrl = "ws://host.docker.internal:8080/ws";
  // process.env.NEXT_PUBLIC_STAGE === "production"
  //   ? "ws://data_ingest:8080/ws"
  //   : "ws://localhost:8080/ws";

  const wss2Url = "ws://host.docker.internal:8090/ws";
  // process.env.NEXT_PUBLIC_STAGE === "production"
  //   ? "ws://trading_algo:8090/ws"
  //   : "ws://localhost:8090/ws";

  console.log(wssUrl, wss2Url);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const socket = new WebSocket(wssUrl);
      const socket2 = new WebSocket(wss2Url);

      socket.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      // console.log("hi");
      socket.onmessage = (event) => {
        const newData = JSON.parse(event.data);
        if (newData.e === "trade") {
          var date = new Date(newData.E * 1000);

          // Extract hours, minutes, and seconds
          const hours = String(date.getHours()).padStart(2, "0");
          const minutes = String(date.getMinutes()).padStart(2, "0");
          const seconds = String(date.getSeconds()).padStart(2, "0");
          newData.date = hours + ":" + minutes + ":" + seconds;
          // console.log(newData);
          // Add to data if it's a kline event, regardless of "x" status
          setData((prevData) => [...prevData, newData]);
        } else if (newData.e === "kline") {
          if (newData.k.x === true) {
            setSeries((prevSeries) => [
              {
                data: [
                  ...prevSeries[0].data,
                  {
                    x: new Date(newData.E), // Convert UNIX timestamp to JavaScript Date object
                    y: [
                      parseFloat(newData.k.o), // open
                      parseFloat(newData.k.h), // high
                      parseFloat(newData.k.l), // low
                      parseFloat(newData.k.c), // close
                    ],
                  },
                ],
              },
            ]);
          }
        }
      };

      socket2.onmessage = (event) => {
        const newData = JSON.parse(event.data);
        console.log("lol");
        console.log(newData);
        setEma(newData);
      };

      return () => {
        socket.close();
      };
    }
  }, []);

  const options: ApexOptions = {
    chart: {
      type: "candlestick", // Use 'candlestick' instead of 'string'
      height: 350,
    },
    // title: {
    //   text: "CandleStick Chart",
    //   align: "left",
    // },
    xaxis: {
      type: "datetime", // Correctly specified as 'datetime'
    },
    yaxis: {
      tooltip: {
        enabled: true,
      },
    },
  };

  return (
    <Box display={"flex"} flexDirection={"column"} gap={"20px"}>
      <Box
        display={"flex"}
        alignContent={"center"}
        flexDirection={tablet ? "row" : "column"}
        gap={"5%"}
        height={"380px"}
      >
        <Box
          width={tablet ? "50%" : "100%"}
          sx={{ border: 1, borderRadius: "25px" }}
        >
          <ReactApexChart
            options={options}
            series={series}
            type="candlestick"
            height={350}
          />
        </Box>

        {/* Render data */}
        <ScrollableBox
          padding={"10px"}
          width={tablet ? "50%" : "100%"}
          overflow={"scroll"}
          alignSelf={"center"}
          customSx={{ border: 1, borderRadius: "25px" }}
        >
          {data.length > 0 ? (
            data
              .slice()
              .reverse()
              .map((item, index) => (
                <p style={{ fontSize: 12 }} key={index}>
                  {item.date}: {item.s} was bought for {item.p} at volume{" "}
                  {item.q}
                </p>
              ))
          ) : (
            <p>Waiting for data from websocket...</p>
          )}
        </ScrollableBox>
      </Box>
      <ScrollableBox
        padding={"10px"}
        minHeight={"100px"}
        width={"100%"}
        customSx={{ border: 1, borderRadius: "25px" }}
      >
        {ema.length > 0 ? (
          <div>
            <p>EMA Values:</p>
            {ema
              .slice()
              .reverse()
              .map((item: string, idx: number) => (
                <p key={idx}>{item}</p>
              ))}
          </div>
        ) : (
          <p>Waiting for data from websocket...</p>
        )}
      </ScrollableBox>
    </Box>
  );
};

export default WebSocketComponent;
