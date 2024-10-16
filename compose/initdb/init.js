try {
  rs.status()
} catch (error) {
  switch (error.codeName) {
    case 'NoReplicationEnabled':
      break;
    case 'NotYetInitialized':
      rs.initiate({ _id: "mender", members: [{ _id: 0, host: '127.0.0.1:27017' }] })
      break;
    default:
      console.error(`error ${error.codeName} (${error.code}) unexpected: terminating`);
      process.exit(1);
  }
}
