package peers

// loadConfig signing key from/to a JSON encoded file
// func (ab *AddressBook) loadConfig(configPath string) error {
// 	ctx := context.Background()
// 	peerPath := filepath.Join(configPath, "config.json")
// 	if _, err := os.Stat(peerPath); err == nil {
// 		cfg, err := loadConfig(peerPath)
// 		if err != nil {
// 			return err
// 		}
// 		keyBytes, err := base58.Decode(cfg.Key)
// 		if err != nil {
// 			return err
// 		}
// 		o, err := encoding.Unmarshal(keyBytes)
// 		if err != nil {
// 			return err
// 		}
// 		ab.localKey = &crypto.Key{}
// 		return ab.localKey.FromObject(o)
// 	}

// 	logger := log.Logger(ctx)
// 	logger.Info("* Configs do not exist, creating new ones.")

// 	peerSigningKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		return err
// 	}

// 	localKey, err := crypto.NewKey(peerSigningKey)
// 	if err != nil {
// 		return err
// 	}

// 	ab.localKey = localKey

// 	keyBytes, err := encoding.Marshal(localKey.ToObject())
// 	if err != nil {
// 		return err
// 	}

// 	cfg := &config{
// 		Key: base58.Encode(keyBytes),
// 	}

// 	if err := storeConfig(cfg, peerPath); err != nil {
// 		return err
// 	}

// 	return nil
// }

// type config struct {
// 	Key string `json:"key"`
// }

// // loadConfig from a JSON encoded file
// func loadConfig(path string) (*config, error) {
// 	bytes, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cfg := &config{}
// 	if err := json.Unmarshal(bytes, cfg); err != nil {
// 		return nil, err
// 	}

// 	return cfg, nil
// }

// // storeConfig to a JSON encoded file
// func storeConfig(cfg *config, path string) error {
// 	bc, _ := json.MarshalIndent(cfg, "", "  ")
// 	return ioutil.WriteFile(path, bc, 0644)
// }
