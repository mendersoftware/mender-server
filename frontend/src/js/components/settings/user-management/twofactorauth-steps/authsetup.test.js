// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import React from 'react';

import { defaultState, undefineds } from '../../../../../../tests/mockData';
import { render } from '../../../../../../tests/setupTests';
import AuthSetup from './authsetup';

const qr =
  'iVBORw0KGgoAAAANSUhEUgAAAYgAAAGIAQAAAABzOEqLAAAAJHRFWHRTb2Z0d2FyZQBRUi1QTkcgaHR0cDovL3FyLnN3dGNoLmNvbS9nj329AAAJe0lEQVR4AWP4TyoY1TGqY1THqI5RHQgdI1cHw/////8zMDAwMDAw/GdgYPjPwMDA8J/h////DAz/GRj+M/z/z8Dw/z8DAwMDAwPD/////4/qoKUOhv////9n+P//////DP8Z/jP8/////////xn+/////z/DfwYGBgYGBob/DP//////n+H/////R3XQUgfD/////zP8Z2Bg+M/wn4Hh/3+G/wz///9nYPjP8P8/AwPD/////zP8/8/wn4GB4T/D/////4/qoKUOhv////9n+M/AwPCf4T/Df4b//xn+MzAw/Gf4z8Dw/z8DA8P//wz/Gf4z/GdgYPjP8P////+jOmipg+H/////Gf4zMDD8Z/jPwMDwn4GBgYGBgYHhPwMDw///DAz/Gf4z/P/P8J+BgeE/w/////+P6qClDob/////Z/j//////wz//zP8/////38GBob/DAwM////Z/jPwPCfgYHhP8P//////2f4/////1EdtNTB8P////8MDAwMDAwM/xn+M/xn+M/wn+E/w3+G/wz/Gf4z/Gf4z/Cf4T8DAwMDAwPD/////4/qoKUOhv////////////////9nYGBg+P///38Ghv8M//////+f4T8DAwMDA8P/////////////////j+qgpQ6G///////PwPCf4T8Dw38GBob/DP8ZGBj+////n4HhPwMDw38Ghv8M/xn+MzAwMDD8/////6gOWupg+P///////xkYGBj+///PwPCf4T/Df4b/////Z2BgYPj//z8Dw////xkY/v//z8DA8P////+jOmipg+H/////GRgY/v9nYGBg+M/w/z/Df4b//xkYGP7/Z/jP8P8/AwPDf4b/DP//M/xnYPj/////UR201MHw//////8ZGBgYGP4z/P/PwPD//38Ghv8M/xn+MzAwMDAwMDD8////P8N/hv//Gf7//////6gOWupg+P//////DAz//zMwMPz//5/hP8P//wwM/xn+M/z/z/D/////DAwMDAwM//8z/Gdg+P////9RHbTUwfD//////xkYGP4z/P/PwMDAwMDA8J/hP8N/hv8MDAwM/xn+////n4GB4f//////M/z/////qA5a6mD4////fwaG/wwM/xn+M/z/z8DAwPCfgeH//////zMwMDD8//+f4T8DAwMDAwMDA8P/////j+qgpQ6G/////////z/D////Gf4zMPxnYGD4z8DA8P8/A8N/Bob/DP8ZGP7/////P8P///////////+oDlrqYPj//////wwM////Z2D4////f4b/DAwMDAz/////z8DAwPCfgeH///8MDAz/Gf4z/P//////UR201MHw////////M/xn+P+f4T/D//////9nYGBg+P+fgeE/AwPDfwYGhv///zP8/8/wn4Hh/////0d10FIHw////////8/AwPCfgeE/AwPDfwYGBob/DP8Z/jP8//+f4f///////2dgYGD4z8DA8P////+jOmipg+H//////zP8/8/A8P8/w///DAz/Gf7//////38GBgYGBgYGBob/DAz//zP8//+f4f//////j+qgpQ6G/////////////wwM/////8/AwPD//38Ghv8MDAz/Gf4z/GdgYGBgYGBg+P//PwPD/////4/qoKUOhv////9n+P//PwPDfwaG/wz/Gf4zMDD8/8/A8P//f4b/DP8Z/jP8Z2Bg+P///38Ghv////8f1UFLHQz/////z/CfgeE/AwMDA8N/BgYGBob/DP8ZGP7//8/A8P//f4b//xn+M/z/z8DAwPD/////ozpoqYPh/////xn+/2f4z/D/PwMDw///DAwMDP///////z/DfwYGBgYGBgaG/wz//////5/h/////0d10FIHw////////8/A8P8/w///DP8ZGP7///+f4f////8ZGBj+////n+H///8MDP//M/z/z/D/////ozpoqYPh/////xkYGP7///+fgeE/w/////8zMPxnYPj/n+E/A8P//wz/Gf7/Z2Bg+P//PwPD/////4/qoKUOhv////9n+M/w/z8DA8N/hv//Gf4z/Gdg+P///3+G////M/z/////fwYGBgaG/wwMDP////8/qoOWOhj+////n4GBgYGB4T/DfwaG/wwMDAwMDAwM////Z2Bg+P+fgYGB4f/////////PwPD/////ozpoqYPh/////xkYGP7/Z2D4z/Cf4T/Df4b//xkY/v9nYPj/n+E/AwPDf4b/DAz//zP8Z/j//////6M6aKmD4f/////////P8J/h/38GBob//xn+M/z///8/w///DP///2dgYPj/n4GB4f9/hv8MDP////8/qoOWOhj+////n+E/AwMDAwMDw///DP8Z/v///5+B4T8Dw/////8zMDAw/P/PwPD/////DAz/////P6qDljoY/v////8/w38GBob//xn+MzD8Z/j/n+H/fwaG//8Z/jMw/GdgYPj/n+E/AwPD/////////39UBy11MPz///8/w///DP//M/z//5/h/3+G///////PwPD//3+G//8Z/jP8/8/AwMDAwPD/P8P/////j+qgpQ6G/////////////////2f4z8Dw/z8DAwMDAwPD////Gf7//8/wn4GB4f///wz/GRgY/v///39UBy11MPz///8/AwMDAwMDw3+G////MzD8Z2Bg+M/AwPD//3+G///////PwPCf4T/DfwYGhv////8f1UFLHQz/////z/D//////xn+/2f4z/D/PwMDAwPD/////zMw/P///z8Dw3+G////M/z/z8Dw/////6M6aKmD4f////8Z/jMwMPxn+M/A8P8/A8P/////MzAw/P/P8P////8M/xn+MzAwMDD8/8/w//////9HddBSB8P/////M/xnYGD4z/D/////DAwM/xn+/2dg+P+f4T/D////GRj+/2f4/5+BgeE/w///////H9VBSx0M/////8/wn4GB4T/Df4b/DP8ZGBj+MzAwMDD8/////////xn+/2f4//8/A8P//wwM/////z+qg5Y6GP7///+f4f//////M/xnYGBg+M/w/z8DAwPD//8MDAwMDAwMDP8ZGP4zMPxn+M/w//////9HddBSB8P/////MzAwMDAwMPz///8/AwMDAwPDfwYGBob//xkY/v9nYPjPwPD//3+G//8ZGP7///9/VActdTD8JxWM6hjVMapjVMeoDoSOkasDAEN0S0qMdGQGAAAAAElFTkSuQmCC';

describe('AuthSetup Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<AuthSetup currentUser={defaultState.users.byId.a1} handle2FAState={jest.fn} qrImage={qr} verify2FA={jest.fn} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
