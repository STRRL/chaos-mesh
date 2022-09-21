/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

const actions = {
  AWSChaos: ['ec2-stop', 'ec2-restart', 'detach-volume'],
  DNSChaos: ['error', 'random'],
  GCPChaos: ['node-stop', 'node-reset', 'disk-loss'],
  HTTPChaos: [],
  IOChaos: ['latency', 'fault', 'attrOverride', 'mistake'],
  JVMChaos: ['latency', 'return', 'exception', 'stress', 'gc', 'ruleData', 'mysql'],
  KernelChaos: [],
  NetworkChaos: ['netem', 'delay', 'loss', 'duplicate', 'corrupt', 'partition', 'bandwidth'],
  PhysicalMachineChaos: [
    'stress-cpu',
    'stress-mem',
    'disk-read-payload',
    'disk-write-payload',
    'disk-fill',
    'network-corrupt',
    'network-duplicate',
    'network-loss',
    'network-delay',
    'network-partition',
    'network-dns',
    'network-bandwidth',
    'network-flood',
    'network-down',
    'process',
    'jvm-exception',
    'jvm-gc',
    'jvm-latency',
    'jvm-return',
    'jvm-stress',
    'jvm-rule-data',
    'jvm-mysql',
    'clock',
    'redis-expiration',
    'redis-penetration',
    'redis-cacheLimit',
    'redis-restart',
    'redis-stop',
    'kafka-fill',
    'kafka-flood',
    'kafka-io',
    'file-create',
    'file-modify',
    'file-delete',
    'file-rename',
    'file-append',
    'file-replace',
    'vm',
    'user_defined',
  ],
  PodChaos: ['pod-kill', 'pod-failure', 'container-kill'],
  StressChaos: [],
  TimeChaos: [],
}

export default actions
