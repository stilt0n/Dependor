import { useState, useEffect, FC } from "react";

const ComponentWithDynamicData: FC<> = () => {
  const [dynamicData, setDynamicData] = useState<any>();
  useEffect(() => {
    const dynamicallyLoadFunc = async () => {
      const d = await import("dynamic_data");
      setDynamicData(d);
    };
    dynamicallyLoadFunc();
  }, []);

  console.log(dynamicData?.getData() ?? []);
  return <p>Some jsx that should be ignored in the future</p>;
};

export default ComponentWithDynamicData;
