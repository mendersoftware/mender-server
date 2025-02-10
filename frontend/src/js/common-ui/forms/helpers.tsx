import {FormProvider, useForm} from "react-hook-form";
import {render} from "../../../../tests/setupTests";
import React from "react";

export const formRenderWrapper = ui => {
  const Wrapper = ({ children }) => {
    const methods = useForm();
    return <FormProvider {...methods}>{children}</FormProvider>;
  };
  return render(<Wrapper>{ui}</Wrapper>);
};
