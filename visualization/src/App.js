import React from 'react';
import logo from './logo.svg';
import './App.css';
import LeafletMap from './components/LeafletMap.js';
import TimeController from './components/TimeController';


/**
 * Global simulation settings. Has to be the same as in Go sources.
 * TODO - refactor to keep in 1 place (maybe in simulation config)
 */
export const Settings = {
  
  SamplingPeriod_ms: 10,

};

class App extends React.Component {

  constructor(props) {
    super(props);
    this.state = {
      runSimulationRev: null,
      simulationStop_ms: null
    }
  }

  runSimulation(timestamp, simulationStop_ms) {
    let rev = this.state.runSimulationRev ? this.state.runSimulationRev + 1 : 1;
    
    this.setState({
      runSimulationRev: rev,
      timestamp: timestamp,
      simulationStop_ms: simulationStop_ms
    })
  }

  render() {
    return (
      <div>
        <LeafletMap runSimulationRev={this.state.runSimulationRev} timestamp={this.state.timestamp} simulationStop_ms={this.state.simulationStop_ms}/>
        <TimeController notifyStartSimulation={(ts, simulationStop_ms) => this.runSimulation(ts, simulationStop_ms)} />
      </div>
    );
  }
}


export default App;
