/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow strict-local
 * @format
 */
import type {promql_return_object} from '@fbcnms/magma-api';

import 'jest-dom/extend-expect';
import EnodebDetail from '../EnodebDetailMain';
import MagmaAPIBindings from '@fbcnms/magma-api';
import MomentUtils from '@date-io/moment';
import MuiStylesThemeProvider from '@material-ui/styles/ThemeProvider';
import React from 'react';
import axiosMock from 'axios';
import defaultTheme from '@fbcnms/ui/theme/default';

import {MemoryRouter, Route} from 'react-router-dom';
import {MuiPickersUtilsProvider} from '@material-ui/pickers';
import {MuiThemeProvider} from '@material-ui/core/styles';
import {cleanup, render, wait} from '@testing-library/react';

jest.mock('axios');
jest.mock('@fbcnms/magma-api');
jest.mock('@fbcnms/ui/hooks/useSnackbar');
afterEach(cleanup);

const mockThroughput: promql_return_object = {
  status: 'success',
  data: {
    resultType: 'matrix',
    result: [
      {
        metric: {},
        values: [['1588898968.042', '6']],
      },
    ],
  },
};

const currTime = Date.now();

describe('<Enodeb />', () => {
  beforeEach(() => {
    // eslint-disable-next-line max-len
    MagmaAPIBindings.getNetworksByNetworkIdPrometheusQueryRange.mockResolvedValue(
      mockThroughput,
    );
  });

  afterEach(() => {
    axiosMock.get.mockClear();
  });

  const enbInfo0 = {
    enb: {
      attached_gateway_id: 'us_baltic_gw1',
      config: {
        bandwidth_mhz: 20,
        cell_id: 1,
        device_class: 'Baicells ID TDD/FDD',
        earfcndl: 44290,
        pci: 36,
        special_subframe_pattern: 7,
        subframe_assignment: 2,
        tac: 1,
        transmit_enabled: true,
      },
      name: 'testEnodeb0',
      serial: 'testEnodebSerial0',
    },
    enb_state: {
      enodeb_configured: true,
      enodeb_connected: true,
      fsm_state: 'Completed provisioning eNB. Awaiting new Inform.',
      gps_connected: true,
      gps_latitude: '41.799182',
      gps_longitude: '-88.097308',
      mme_connected: false,
      opstate_enabled: false,
      ptp_connected: false,
      reporting_gateway_id: 'us_baltic_gw1',
      rf_tx_desired: true,
      rf_tx_on: false,
      time_reported: 0,
    },
  };

  const enbInfo1 = Object.assign({}, enbInfo0);
  enbInfo1.enb = {...enbInfo1.enb, name: 'testEnodeb1'};
  enbInfo1.enb_state = {
    ...enbInfo1.enb_state,
    fsm_state: 'initializing',
    time_reported: currTime,
    rf_tx_on: true,
  };
  const enbInfo = {
    testEnodebSerial0: enbInfo0,
    testEnodebSerial1: enbInfo1,
  };

  const Wrapper = () => (
    <MemoryRouter
      initialEntries={['/nms/mynetwork/enodeb/testEnodebSerial0/overview']}
      initialIndex={0}>
      <MuiPickersUtilsProvider utils={MomentUtils}>
        <MuiThemeProvider theme={defaultTheme}>
          <MuiStylesThemeProvider theme={defaultTheme}>
            <Route
              path="/nms/:networkId/enodeb/:enodebSerial/overview"
              render={props => <EnodebDetail {...props} enbInfo={enbInfo} />}
            />
          </MuiStylesThemeProvider>
        </MuiThemeProvider>
      </MuiPickersUtilsProvider>
    </MemoryRouter>
  );

  it('renders', async () => {
    const {getByTestId} = render(<Wrapper />);
    await wait();

    // TODO - commenting this out till we have per enodeb metric support
    // expect(
    //   MagmaAPIBindings.getNetworksByNetworkIdPrometheusQueryRange,
    // ).toHaveBeenCalledTimes(2);
    expect(getByTestId('eNodeB Serial Number')).toHaveTextContent(
      'testEnodebSerial0',
    );
    expect(getByTestId('Health')).toHaveTextContent('Bad');
    expect(getByTestId('Transmit Enabled')).toHaveTextContent('Enabled');
    expect(getByTestId('Gateway ID')).toHaveTextContent('us_baltic_gw1');
    expect(getByTestId('Mme Connected')).toHaveTextContent('Disconnected');
  });
});