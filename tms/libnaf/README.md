# Introduction

The North Atlantic Format (NAF) is used for fisheries related electronic data transmission. Some states and regional fisheries management organizations including the North East Atlantic Fisheries commission (NEAFC) and the Northwest Atlantic Fisheries Organization (NAFO) are using NAF elements or messages. In April 2005 the NAF website was deployed (http://www.naf-format.org) to help future users to better understand the standard and contribute to further improvement and expansion. The management of NAF website is a joint responsibility of NAFO, and NEAFC secretariats.

This library encodes and decodes messages based on the NAF format and Mcmurdo’s Omnicom VMS messages standard content. It is intended to enable duplex communication between the Omnicom server and 3rd party modules. The library was developped to answer functional requirements for Omnicom-NAF messages that will be used for VMS operations.


## functions

The library has two exportal functions:

- func ParseNaf(str string) (*iridium.Iridium, error)

- func EncodeNaf(activity *tms.MessageActivity) (string, error)

## bugs

The library has a known bug related to VIN (Volatage Input) in Single position reports. VIN is always set to 0 for single position reports because the customer requires the existance of that field to decode the data, VIN = 0 in export Single Position Report messages does not reflect a real value. The real VIN value for a beacon is present in History Report messages. 
