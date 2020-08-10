package interactor

import (
  "github.com/arstevens/goback/processor"
  "github.com/arstevens/goback/reflector"
  "testing"
)

func TestInteractor(t *testing.T) {
  refTypes := map[processor.ReflectorCode]ReflectorCreator{
    "rf1":reflector.NewPlainReflector,
  }
  cmLoaders := map[processor.ChangeMapCode]ChangeMapLoader{
    "cm1":reflector.LoadSHA1ChangeMap,
  }
  cmCreator := map[processor.ChangeMapCode]ChangeMapCreator{
    "cm1":reflector.NewSHA1ChangeMap,
  }

  g := ReflectionGenerator{
    reflectorTypes:refTypes,
    changeMapCreators:cmCreator,
    changeMapLoaders:cmLoaders,
  }

  serial := `0,testzone,CbOue9P6wkR2SyeceD28O0ippzs=,true)3,graphic_mockups,UjYLX8f9kyEAnjfsX4SG3n+Ef4A=,true)4,honeycoin_load_change.drawio,z3s12tTqZEgwoz1jNia3B4oksYM=,false)|)5,honeycoin_mining_v1.drawio,awVf4lhtZtpqyr790J+9trtwwu4=,false)|)6,load_distribution_v1.drawio,HHkH1VueYC7aYUDkTjMlTTulLbc=,false)|)7,load_pickup_v1.drawio,M031xhXfr9laZ75H6ZRwWq6bAKU=,false)|)8,simple_exchange_v1.drawio,z8ld9Lfx3MbdKa34RelzzMLP/0k=,false)|)9,topology_graph_v1.drawio,kO9QBq3i9QDm2wQb9ENkr+NkLSU=,false)|)|)10,graphics,zuiDznL+4eDuCYzwSFmhkWQ2xtg=,true)11,honeycoin_load_change.png,/ow/GtiyJo3oh1r0wpcCkPWaF54=,false)|)12,honeycoin_mining_v1.png,7PABOxbIMhYTHg9RNEEBcugazW4=,false)|)13,load_distribution_v1.png,JgzHHoLznr1pW4huCV71J7hoOdY=,false)|)14,load_pickup_v1.png,UfljFkCsLmCMLOXQG0kUkw6AWkc=,false)|)15,simple_exchange_v1.png,eSj6Zmp8nBkml15LzTAuEBpggmU=,false)|)16,topology_graph_v1.png,eLdHhek5e7wj0pC7/1QwyG36rR4=,false)|)|)2,Hive_Whitepaper_v3_Fluid.odt,q23GRVlHATRR7QPhEZPYoVkLwqQ=,false)|)|)`
  cm1, err := g.OpenChangeMap("cm1", "/home/aleksandr/Workspace/", serial)
  if err != nil {
    panic(err)
  }

  cm2, err := g.NewChangeMap("cm1", "/home/aleksandr/Workspace/testzone2")
  if err != nil {
    panic(err)
  }

  ref1, err := g.Reflect("rf1", cm1, cm2)
  if err != nil {
    panic(err)
  }

  err = ref1.Backup()
  if err != nil {
    panic(err)
  }


}
