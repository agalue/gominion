package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/broker"
	"github.com/agalue/gominion/collectors"
	"github.com/agalue/gominion/detectors"
	"github.com/agalue/gominion/monitors"
	_ "github.com/agalue/gominion/rpc"  // Load all RPC modules
	_ "github.com/agalue/gominion/sink" // Load all Sink modules

	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultLocation = "Local"
const defaultBroker = "localhost:8990"
const defaultTrapPort = 1162
const defaultSyslogPort = 1514

var cfgFile string
var minionConfig = &api.MinionConfig{BrokerType: "grpc"}
var client = &broker.GrpcClient{}

// rootCmd represents the base command that starts the Minion's gRPC client
var rootCmd = &cobra.Command{
	Use:     "gominion",
	Short:   "An implementation of OpenNMS Minion in Go",
	Version: "0.1.0-alpha1",
	Run:     rootHandler,
}

// Execute prepares the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Initialize Flags
	hostname, _ := os.Hostname()
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is gominion.yaml)")
	rootCmd.Flags().StringVarP(&minionConfig.ID, "id", "i", "", fmt.Sprintf("Minion ID (default is %s)", hostname))
	rootCmd.Flags().StringVarP(&minionConfig.Location, "location", "l", "", fmt.Sprintf("Minion Location (default is %s)", defaultLocation))
	rootCmd.Flags().StringVarP(&minionConfig.BrokerURL, "broker", "b", "", fmt.Sprintf("Broker URL (default is %s)", defaultBroker))
	rootCmd.Flags().IntVarP(&minionConfig.TrapPort, "trapPort", "t", 0, fmt.Sprintf("SNMP Trap port (default is %d)", defaultTrapPort))
	rootCmd.Flags().IntVarP(&minionConfig.SyslogPort, "syslogPort", "s", 0, fmt.Sprintf("Syslog port (default is %d)", defaultSyslogPort))
	rootCmd.Flags().IntVarP(&minionConfig.NxosGrpcPort, "nxosGrpcPort", "n", 0, "Cisco NX-OS gRPC port")

	// Initialize Flag Binding
	viper.BindPFlag("id", rootCmd.Flags().Lookup("id"))
	viper.BindPFlag("location", rootCmd.Flags().Lookup("location"))
	viper.BindPFlag("broker", rootCmd.Flags().Lookup("broker"))
	viper.BindPFlag("trapPort", rootCmd.Flags().Lookup("trapPort"))
	viper.BindPFlag("syslogPort", rootCmd.Flags().Lookup("syslogPort"))
	viper.BindPFlag("nxosGrpcPort", rootCmd.Flags().Lookup("nxosGrpcPort"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "gominion" (without extension).
		if home, err := homedir.Dir(); err != nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName(".gominion")
	}

	viper.SetEnvPrefix("GOMINION")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Initialize configuration, applying defaults on missing attributes
	hostname, _ := os.Hostname()
	minionConfig.ID = getString("id", hostname)
	minionConfig.Location = getString("location", defaultLocation)
	minionConfig.BrokerURL = getString("broker", defaultBroker)
	minionConfig.TrapPort = getInt("trapPort", defaultTrapPort)
	minionConfig.SyslogPort = getInt("syslogPort", defaultSyslogPort)
	minionConfig.NxosGrpcPort = getInt("nxosGrpcPort", 0)
}

func rootHandler(cmd *cobra.Command, args []string) {
	log.Printf("Starting OpenNMS Minion...\n%s", minionConfig.String())
	if err := minionConfig.IsValid(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	for _, m := range api.GetAllRPCModules() {
		log.Printf("Registered RPC module %s", m.GetID())
	}
	for _, m := range api.GetAllSinkModules() {
		log.Printf("Registered Sink module %s", m.GetID())
	}
	for _, m := range collectors.GetAllCollectors() {
		log.Printf("Registered collector module %s", m.GetID())
	}
	for _, m := range detectors.GetAllDetectors() {
		log.Printf("Registered detector module %s", m.GetID())
	}
	for _, m := range monitors.GetAllMonitors() {
		log.Printf("Registered poller module %s", m.GetID())
	}
	if err := client.Start(minionConfig); err != nil {
		log.Fatalf("Cannot connect to OpenNMS gRPC server: %v", err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	client.Stop()
}

func getString(key string, defaultValue string) string {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value := viper.GetInt(key)
	if value == 0 {
		return defaultValue
	}
	return value
}
